package http

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/atproto"
	"github.com/simbachu/twisky/internal/bluesky"
	composepage "github.com/simbachu/twisky/internal/components/compose"
	loginpage "github.com/simbachu/twisky/internal/components/login"
	postpage "github.com/simbachu/twisky/internal/components/post"
	"github.com/simbachu/twisky/internal/intent"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func (s *Server) handleComposeNew(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}
	if _, err := s.loadActiveAccount(r); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		returnPath := "/my/posts/new"
		if parent := strings.TrimSpace(r.URL.Query().Get("parent")); parent != "" {
			returnPath = "/my/posts/new?parent=" + url.QueryEscape(parent)
		}
		_ = loginpage.Page("", s.publicBaseURL, returnPath, false).Render(w)
		return
	}
	view := s.composePageView(w, r, "", "")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = composepage.NewPostPage(view).Render(w)
}

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	text := r.FormValue("text")
	parentURI := strings.TrimSpace(r.FormValue("parent"))

	account, err := s.loadActiveAccount(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	writer := s.postWriter
	fetcher := s.postFetcher
	if writer == nil {
		_, client, err := s.ResumeActiveClient(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		writer = client
		if fetcher == nil {
			fetcher = client
		}
	}
	if parentURI != "" && fetcher == nil {
		if client := s.sessionClient(r); client != nil {
			fetcher = client
		}
	}
	if s.commands == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	create := intent.CreatePost{Text: text}
	var parentPost *bluesky.Post
	if parentURI != "" {
		reply, parent, err := resolveReplyTo(fetcher, r.Context(), parentURI)
		parentPost = parent
		if err != nil || reply == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = composepage.NewPostPage(s.composePageFromParent(w, r, "Could not load the post to reply to.", text, "", nil)).Render(w)
			return
		}
		create.Reply = reply
	}

	recordURI, err := s.commands.Dispatch(r.Context(), create, writer)
	if err != nil {
		message := composeErrorMessage(err)
		if message != "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			view := s.composePageFromParent(w, r, message, text, parentURI, parentPost)
			_ = composepage.NewPostPage(view).Render(w)
			return
		}
		http.Error(w, "post failed", http.StatusBadGateway)
		return
	}

	parsed, err := atproto.ParsePostURI(recordURI)
	if err != nil {
		http.Error(w, "post failed", http.StatusBadGateway)
		return
	}
	location := actor.PostPath(account.Handle, account.DID, parsed.Rkey())
	http.Redirect(w, r, location, http.StatusSeeOther)
}

func (s *Server) composePageView(w http.ResponseWriter, r *http.Request, errorMessage, text string) composepage.PageView {
	parentURI := strings.TrimSpace(r.URL.Query().Get("parent"))
	if parentURI == "" {
		suggested := s.suggestedAccounts(r.Context())
		return composepage.PageView{
			Error:         errorMessage,
			Text:          text,
			PublicBaseURL: s.publicBaseURL,
			Suggested:     suggested,
			Accounts:      s.accountMenuView(w, r, suggested...),
		}
	}
	fetcher := s.postFetcher
	if fetcher == nil {
		if client := s.sessionClient(r); client != nil {
			fetcher = client
		}
	}
	parentPost, err := loadParentPost(fetcher, r.Context(), parentURI)
	message := errorMessage
	if err != nil || parentPost == nil {
		if message == "" {
			message = "Could not load the post to reply to."
		}
		return s.composePageFromParent(w, r, message, text, "", nil)
	}
	return s.composePageFromParent(w, r, message, text, parentURI, parentPost)
}

func (s *Server) composePageFromParent(w http.ResponseWriter, r *http.Request, errorMessage, text, parentURI string, parent *bluesky.Post) composepage.PageView {
	suggested := s.suggestedAccounts(r.Context())
	view := composepage.PageView{
		Error:         errorMessage,
		Text:          text,
		ParentURI:     parentURI,
		PublicBaseURL: s.publicBaseURL,
		Suggested:     suggested,
		Accounts:      s.accountMenuView(w, r, suggested...),
	}
	if parent != nil && parentURI != "" {
		postView := feedquery.InsetPostView(feedquery.NewPostView(*parent))
		view.Parent = postpage.InsetPost(&postView, time.Now().UTC())
	}
	return view
}

type postFetcher interface {
	GetPosts(ctx context.Context, uris []string) ([]bluesky.Post, error)
}

func loadParentPost(fetcher postFetcher, ctx context.Context, parentURI string) (*bluesky.Post, error) {
	if _, err := atproto.ParsePostURI(parentURI); err != nil {
		return nil, err
	}
	if fetcher == nil {
		return nil, errNoPostFetcher
	}
	posts, err := fetcher.GetPosts(ctx, []string{parentURI})
	if err != nil {
		return nil, err
	}
	for i := range posts {
		if posts[i].URI == parentURI {
			return &posts[i], nil
		}
	}
	return nil, errParentNotFound
}

func resolveReplyTo(fetcher postFetcher, ctx context.Context, parentURI string) (*intent.ReplyTo, *bluesky.Post, error) {
	parent, err := loadParentPost(fetcher, ctx, parentURI)
	if err != nil || parent == nil {
		return nil, parent, err
	}
	if parent.CID == "" {
		return nil, parent, errParentNotFound
	}
	reply := ReplyToFromPost(*parent)
	return &reply, parent, nil
}

// ReplyToFromPost derives AT Protocol reply refs from the post being replied to.
func ReplyToFromPost(parent bluesky.Post) intent.ReplyTo {
	rootURI, rootCID := parent.URI, parent.CID
	if parent.Record.Reply != nil && parent.Record.Reply.Root.URI != "" && parent.Record.Reply.Root.CID != "" {
		rootURI = parent.Record.Reply.Root.URI
		rootCID = parent.Record.Reply.Root.CID
	}
	return intent.ReplyTo{
		RootURI:   rootURI,
		RootCID:   rootCID,
		ParentURI: parent.URI,
		ParentCID: parent.CID,
	}
}

var (
	errNoPostFetcher  = errString("compose: post fetcher is required")
	errParentNotFound = errString("compose: parent post not found")
)

type errString string

func (e errString) Error() string { return string(e) }

func composeErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	switch {
	case strings.Contains(msg, "text is required"):
		return "Post text is required."
	case strings.Contains(msg, "exceeds"):
		return "Post text is too long."
	case strings.Contains(msg, "reply refs"):
		return "Could not load the post to reply to."
	default:
		return ""
	}
}
