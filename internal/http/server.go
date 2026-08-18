package http

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/simbachu/twisky/internal/actor"
	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/command"
	"github.com/simbachu/twisky/internal/command/like"
	"github.com/simbachu/twisky/internal/command/repost"
	errorpage "github.com/simbachu/twisky/internal/components/errorpage"
	feedcomponent "github.com/simbachu/twisky/internal/components/feed"
	healthzpage "github.com/simbachu/twisky/internal/components/healthz"
	homepage "github.com/simbachu/twisky/internal/components/home"
	loginpage "github.com/simbachu/twisky/internal/components/login"
	postpage "github.com/simbachu/twisky/internal/components/post"
	profilepage "github.com/simbachu/twisky/internal/components/profile"
	tagpage "github.com/simbachu/twisky/internal/components/tag"
	"github.com/simbachu/twisky/internal/components/ui"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/moderation"
	"github.com/simbachu/twisky/internal/query"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	homequery "github.com/simbachu/twisky/internal/query/home"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/suggestions"
	"github.com/simbachu/twisky/internal/query/tag"
	"github.com/simbachu/twisky/internal/response"
	"github.com/simbachu/twisky/internal/version"
	"github.com/simbachu/twisky/static"
)

type Server struct {
	queries       *query.Dispatcher
	commands      *command.Dispatcher
	suggestions   *suggestions.Handler
	publicBaseURL string
	auth          *authoauth.Service
	prefs         moderation.PrefsProvider
	// likeWriter, when set, skips OAuth resume for LikePost (tests).
	likeWriter like.Writer
	// repostWriter, when set, skips OAuth resume for RepostPost (tests).
	repostWriter repost.Writer
	// homeReader, when set, is used for the home timeline instead of the session client (tests).
	homeReader homequery.Reader
	// sessionClients caches resumed OAuth API clients briefly to cut ResumeSession
	// traffic from authenticated poll/enrichment paths.
	sessionClients *sessionClientCache
}

func NewServer(queries *query.Dispatcher, commands *command.Dispatcher, suggestionsHandler *suggestions.Handler, publicBaseURL string, auth *authoauth.Service) *Server {
	return &Server{
		queries:        queries,
		commands:       commands,
		suggestions:    suggestionsHandler,
		publicBaseURL:  publicBaseURL,
		auth:           auth,
		prefs:          moderation.DefaultPrefsProvider{},
		sessionClients: newSessionClientCache(),
	}
}

// WithLikeWriter overrides the OAuth session client used for LikePost (tests).
func (s *Server) WithLikeWriter(w like.Writer) *Server {
	s.likeWriter = w
	return s
}

// WithRepostWriter overrides the OAuth session client used for RepostPost (tests).
func (s *Server) WithRepostWriter(w repost.Writer) *Server {
	s.repostWriter = w
	return s
}

// WithHomeReader overrides the session client used for the home timeline (tests).
func (s *Server) WithHomeReader(r homequery.Reader) *Server {
	s.homeReader = r
	return s
}

func (s *Server) suggestedAccounts(ctx context.Context) []ui.AuthorInfo {
	if s.suggestions == nil {
		return nil
	}
	accounts := s.suggestions.SuggestedAccounts(ctx)
	if len(accounts) == 0 {
		return nil
	}
	authors := make([]ui.AuthorInfo, len(accounts))
	for i, account := range accounts {
		authors[i] = ui.AuthorInfo{
			Handle:      account.Handle,
			DisplayName: account.DisplayName,
			DID:         account.DID,
			Avatar:      account.Avatar,
			IsLabeler:   actor.IsLabelerAccount(account.Handle, account.DID, account.IsLabeler),
		}
	}
	return authors
}

func (s *Server) Handler() http.Handler {
	staticFS, err := fs.Sub(static.WebFS, "web")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServerFS(staticFS)))
	r.Get("/healthz", s.handleHealthz)
	r.Get("/oauth/client-metadata.json", s.handleClientMetadata)
	r.Get("/oauth/login", s.handleOAuthLogin)
	r.Post("/oauth/login", s.handleOAuthLogin)
	r.Get("/oauth/callback", s.handleOAuthCallback)
	r.Post("/oauth/logout", s.handleOAuthLogout)
	r.Post("/action/like", s.handleLike)
	r.Post("/action/repost", s.handleRepost)
	r.Get("/", s.handleHome)
	r.Get("/feed/{feedSlug}", s.handleHome)
	r.Get("/tagged/{tag}", s.handleTag)
	r.Get("/{slug}/post/{id}", s.handlePost)
	r.Get("/{slug}/media", s.handleProfile(intent.ProfileTabMedia))
	r.Get("/{slug}", s.handleProfile(intent.ProfileTabPosts))
	return r
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}

	reader, ok := s.homeTimelineReader(r)
	if !ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = loginpage.Page("", s.publicBaseURL, "/").Render(w)
		return
	}

	cursor, since, refresh := feedFragmentParams(r)
	resp := homequery.NewHandler(reader, s.prefs).Handle(r.Context(), intent.ViewHome{
		Cursor:    cursor,
		FeedSlug:  chi.URLParam(r, "feedSlug"),
		HeadCheck: since != "" && refresh == "" && cursor == "",
	})
	switch v := resp.(type) {
	case homequery.HomeView:
		if since == "" {
			s.enrichFeedLikes(r, &v.Feed)
		}
		if renderFeedFragment(w, r, v.Feed, cursor, since, refresh) {
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		suggested := s.suggestedAccounts(r.Context())
		_ = homepage.Home(v, time.Now().UTC(), suggested, s.accountMenuView(w, r, suggested...), s.publicBaseURL).Render(w)
	case response.ErrorResponse:
		s.writeQueryError(w, r, v)
	default:
		s.writeQueryError(w, r, response.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Something went wrong loading this page",
		})
	}
}

// homeTimelineReader returns the timeline reader for a logged-in request.
// ok is false when there is no active session (caller should render login).
func (s *Server) homeTimelineReader(r *http.Request) (homequery.Reader, bool) {
	if s.homeReader != nil {
		if _, err := s.loadActiveAccount(r); err != nil {
			return nil, false
		}
		return s.homeReader, true
	}
	_, client, err := s.ResumeActiveClient(r)
	if err != nil {
		return nil, false
	}
	return client, true
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if acceptsHTML(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Vary", "Accept")
		_ = healthzpage.Preview(s.publicBaseURL).Render(w)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "ok %s", version.ShortID())
}

func acceptsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

func (s *Server) handleTag(w http.ResponseWriter, r *http.Request) {
	s.dispatchTag(w, r, chi.URLParam(r, "tag"))
}

func (s *Server) dispatchTag(w http.ResponseWriter, r *http.Request, tagName string) {
	cursor, since, refresh := feedFragmentParams(r)
	resp, err := s.queries.Dispatch(r.Context(), intent.ViewTag{
		Tag:       tagName,
		Cursor:    cursor,
		HeadCheck: since != "" && refresh == "" && cursor == "",
	})
	if err != nil {
		s.writeQueryError(w, r, response.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Something went wrong loading this page",
		})
		return
	}

	switch v := resp.(type) {
	case tag.TagView:
		if since == "" {
			s.enrichTagView(r, &v)
		}
		if renderFeedFragment(w, r, v.Feed, cursor, since, refresh) {
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		suggested := s.suggestedAccounts(r.Context())
		_ = tagpage.Tag(v, time.Now().UTC(), suggested, s.accountMenuView(w, r, suggested...), s.publicBaseURL).Render(w)
	case response.ErrorResponse:
		s.writeQueryError(w, r, v)
	default:
		s.writeQueryError(w, r, response.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Something went wrong loading this page",
		})
	}
}

func (s *Server) handleProfile(tab intent.ProfileTab) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		cursor, since, refresh := feedFragmentParams(r)
		resp, err := s.queries.Dispatch(r.Context(), intent.ViewProfile{
			Slug:      slug,
			Tab:       tab,
			Cursor:    cursor,
			HeadCheck: since != "" && refresh == "" && cursor == "",
		})
		if err != nil {
			s.writeQueryError(w, r, response.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Something went wrong loading this page",
			})
			return
		}

		switch v := resp.(type) {
		case profile.ProfileView:
			if since == "" {
				s.enrichProfileView(r, &v)
			}
			if renderFeedFragment(w, r, v.Feed, cursor, since, refresh) {
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			suggested := s.suggestedAccounts(r.Context())
			known := append([]ui.AuthorInfo{authorInfoFromProfileView(v)}, suggested...)
			_ = profilepage.Profile(v, time.Now().UTC(), suggested, s.accountMenuView(w, r, known...), s.publicBaseURL).Render(w)
		case response.ErrorResponse:
			s.writeQueryError(w, r, v)
		default:
			s.writeQueryError(w, r, response.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Something went wrong loading this page",
			})
		}
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	postID, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		s.writeQueryError(w, r, response.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid post identifier",
		})
		return
	}

	resp, err := s.queries.Dispatch(r.Context(), intent.ViewPost{
		Slug: chi.URLParam(r, "slug"),
		ID:   postID,
		Part: postPagePart(r),
	})
	if err != nil {
		s.writeQueryError(w, r, response.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Something went wrong loading this page",
		})
		return
	}

	switch v := resp.(type) {
	case feedquery.PostPageView:
		part := postPagePart(r)
		// Counts fragments only need public engagement numbers; skip OAuth resume
		// and viewer like enrichment on that hot poll path.
		if part != feedquery.PostPagePartCounts {
			s.enrichPageLikes(r, &v)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		now := time.Now().UTC()
		switch part {
		case feedquery.PostPagePartAncestors:
			_ = postpage.PostPageAncestors(v, now).Render(w)
		case feedquery.PostPagePartCounts:
			if live, ok := liveToggleParam(r); ok {
				_ = postpage.CountsToggleFragment(v.Post, now, live).Render(w)
			} else {
				_ = postpage.CountsRefreshFragment(v.Post, previousCounts(r), now, wantsStats(r)).Render(w)
			}
		case feedquery.PostPagePartReplies:
			_ = postpage.RepliesRefreshFragment(v, parseKnownParam(r), now).Render(w)
		default:
			v.ExplicitLive = wantsLive(r)
			suggested := s.suggestedAccounts(r.Context())
			_ = postpage.PostPage(v, now, suggested, s.accountMenuView(w, r, suggested...), s.publicBaseURL).Render(w)
		}
	case response.ErrorResponse:
		s.writeQueryError(w, r, v)
	default:
		s.writeQueryError(w, r, response.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Something went wrong loading this page",
		})
	}
}

func (s *Server) writeQueryError(w http.ResponseWriter, r *http.Request, errResp response.ErrorResponse) {
	status := errResp.Status
	if status == 0 {
		status = http.StatusInternalServerError
	}
	message := strings.TrimSpace(errResp.Message)
	if message == "" {
		message = "Something went wrong loading this page"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if isHTMXRequest(r) {
		w.Header().Set("HX-Reswap", "none")
		w.WriteHeader(status)
		_ = errorpage.AlertFragment(message, status).Render(w)
		return
	}

	w.WriteHeader(status)
	suggested := s.suggestedAccounts(r.Context())
	_ = errorpage.Page(
		message,
		status,
		r.URL.Path,
		suggested,
		s.accountMenuView(w, r, suggested...),
		s.publicBaseURL,
	).Render(w)
}

func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func feedFragmentParams(r *http.Request) (cursor, since, refresh string) {
	query := r.URL.Query()
	return query.Get("cursor"), query.Get("since"), query.Get("refresh")
}

func postPagePart(r *http.Request) string {
	q := r.URL.Query()
	if q.Get("ancestors") == "1" {
		return feedquery.PostPagePartAncestors
	}
	if q.Get("counts") == "1" {
		return feedquery.PostPagePartCounts
	}
	if q.Get("replies") == "1" {
		return feedquery.PostPagePartReplies
	}
	return ""
}

func wantsLive(r *http.Request) bool {
	return r.URL.Query().Get("live") == "1"
}

func wantsStats(r *http.Request) bool {
	return r.URL.Query().Get("stats") == "1"
}

// liveToggleParam reports the requested live state and whether the request
// was a play/pause toggle at all (absent for a periodic refresh poll).
func liveToggleParam(r *http.Request) (live bool, present bool) {
	switch r.URL.Query().Get("live") {
	case "1":
		return true, true
	case "0":
		return false, true
	default:
		return false, false
	}
}

// previousCounts reads the like/reply/repost counts a client currently has
// displayed, as reported back on a periodic counts poll.
func previousCounts(r *http.Request) postpage.PreviousCounts {
	q := r.URL.Query()
	return postpage.PreviousCounts{
		Reply:  parseCountParam(q.Get("reply")),
		Repost: parseCountParam(q.Get("repost")),
		Like:   parseCountParam(q.Get("like")),
	}
}

func parseCountParam(raw string) *int {
	if raw == "" {
		return nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &n
}

// parseKnownParam reads the comma-separated reply post IDs the client already
// has rendered, for the replies live-refresh fragment.
func parseKnownParam(r *http.Request) map[string]bool {
	raw := strings.TrimSpace(r.URL.Query().Get("known"))
	if raw == "" {
		return map[string]bool{}
	}
	known := make(map[string]bool)
	for _, id := range strings.Split(raw, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			known[id] = true
		}
	}
	return known
}

func renderFeedFragment(
	w http.ResponseWriter,
	r *http.Request,
	feed feedquery.FeedView,
	cursor, since, refresh string,
) bool {
	now := time.Now().UTC()
	feedURL := r.URL.Path

	switch {
	case cursor != "":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = feedcomponent.FeedItems(feed, now, feedURL).Render(w)
		return true
	case since != "":
		newPosts := feedquery.NewPostsSince(feed.Posts, since)
		banner := feedcomponent.NewPostsBanner(len(newPosts), feedURL, since)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if banner != nil {
			_ = banner.Render(w)
		}
		return true
	case refresh != "":
		newPosts := feedquery.NewPostsSince(feed.Posts, refresh)
		newTop := refresh
		if len(newPosts) > 0 {
			newTop = newPosts[0].ID
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		nodes := feedcomponent.PrependItems(newPosts, now)
		nodes = append(nodes, feedcomponent.NewPostsPollOOB(feedURL, newTop))
		_ = nodes.Render(w)
		return true
	default:
		return false
	}
}

const shutdownTimeout = 10 * time.Second

func ListenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	server := &http.Server{Addr: addr, Handler: handler}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("http server: %w", err)
	}
}
