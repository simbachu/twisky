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
	feedcomponent "github.com/simbachu/twisky/internal/components/feed"
	postpage "github.com/simbachu/twisky/internal/components/post"
	profilepage "github.com/simbachu/twisky/internal/components/profile"
	tagpage "github.com/simbachu/twisky/internal/components/tag"
	"github.com/simbachu/twisky/internal/components/ui"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/suggestions"
	"github.com/simbachu/twisky/internal/query/tag"
	"github.com/simbachu/twisky/internal/response"
	"github.com/simbachu/twisky/internal/version"
	"github.com/simbachu/twisky/static"
)

type Server struct {
	queries       *query.Dispatcher
	suggestions   *suggestions.Handler
	publicBaseURL string
}

func NewServer(queries *query.Dispatcher, suggestionsHandler *suggestions.Handler, publicBaseURL string) *Server {
	return &Server{queries: queries, suggestions: suggestionsHandler, publicBaseURL: publicBaseURL}
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
			Avatar:      account.Avatar,
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
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServerFS(staticFS)))
	r.Get("/healthz", s.handleHealthz)
	r.Get("/tagged/{tag}", s.handleTag)
	r.Get("/{slug}/post/{id}", s.handlePost)
	r.Get("/{slug}/media", s.handleProfile(intent.ProfileTabMedia))
	r.Get("/{slug}", s.handleProfile(intent.ProfileTabPosts))
	return r
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "ok %s", version.ShortID())
}

func (s *Server) handleTag(w http.ResponseWriter, r *http.Request) {
	s.dispatchTag(w, r, chi.URLParam(r, "tag"))
}

func (s *Server) dispatchTag(w http.ResponseWriter, r *http.Request, tagName string) {
	cursor, since, refresh := feedFragmentParams(r)
	resp, err := s.queries.Dispatch(r.Context(), intent.ViewTag{
		Tag:    tagName,
		Cursor: cursor,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	switch v := resp.(type) {
	case tag.TagView:
		if renderFeedFragment(w, r, v.Feed, cursor, since, refresh) {
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tagpage.Tag(v, time.Now().UTC(), s.suggestedAccounts(r.Context()), s.publicBaseURL).Render(w)
	case response.ErrorResponse:
		http.Error(w, v.Message, v.Status)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (s *Server) handleProfile(tab intent.ProfileTab) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		cursor, since, refresh := feedFragmentParams(r)
		resp, err := s.queries.Dispatch(r.Context(), intent.ViewProfile{
			Slug:   slug,
			Tab:    tab,
			Cursor: cursor,
		})
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		switch v := resp.(type) {
		case profile.ProfileView:
			if renderFeedFragment(w, r, v.Feed, cursor, since, refresh) {
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = profilepage.Profile(v, time.Now().UTC(), s.suggestedAccounts(r.Context()), s.publicBaseURL).Render(w)
		case response.ErrorResponse:
			http.Error(w, v.Message, v.Status)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	postID, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	resp, err := s.queries.Dispatch(r.Context(), intent.ViewPost{
		Slug: chi.URLParam(r, "slug"),
		ID:   postID,
		Part: postPagePart(r),
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	switch v := resp.(type) {
	case feedquery.PostPageView:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		now := time.Now().UTC()
		switch postPagePart(r) {
		case feedquery.PostPagePartAncestors:
			_ = postpage.PostPageAncestors(v, now).Render(w)
		case feedquery.PostPagePartCounts:
			if live, ok := liveToggleParam(r); ok {
				_ = postpage.CountsToggleFragment(v.Post, now, live).Render(w)
			} else {
				_ = postpage.CountsRefreshFragment(v.Post, previousCounts(r), now).Render(w)
			}
		case feedquery.PostPagePartReplies:
			_ = postpage.RepliesRefreshFragment(v, parseKnownParam(r), now).Render(w)
		default:
			v.ExplicitLive = wantsLive(r)
			_ = postpage.PostPage(v, now, s.suggestedAccounts(r.Context()), s.publicBaseURL).Render(w)
		}
	case response.ErrorResponse:
		http.Error(w, v.Message, v.Status)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
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

func ListenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	server := &http.Server{Addr: addr, Handler: handler}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return server.Shutdown(context.Background())
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("http server: %w", err)
	}
}
