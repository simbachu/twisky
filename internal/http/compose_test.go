package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/auth/session"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/command"
	twiskyhttp "github.com/simbachu/twisky/internal/http"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/suggestions"
	"github.com/simbachu/twisky/internal/query/tag"
)

type stubPostWriter struct {
	calls     int
	text      string
	reply     *intent.ReplyTo
	recordURI string
	err       error
}

func (s *stubPostWriter) CreatePost(_ context.Context, text string, reply *intent.ReplyTo) (string, error) {
	s.calls++
	s.text = text
	s.reply = reply
	if s.recordURI == "" {
		s.recordURI = "at://did:plc:alice/app.bsky.feed.post/newpost1"
	}
	return s.recordURI, s.err
}

type stubPostFetcher struct {
	posts []bluesky.Post
	err   error
}

func (s stubPostFetcher) GetPosts(_ context.Context, uris []string) ([]bluesky.Post, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := make([]bluesky.Post, 0, len(uris))
	for _, uri := range uris {
		for _, post := range s.posts {
			if post.URI == uri {
				out = append(out, post)
			}
		}
	}
	return out, nil
}

func newComposeTestServer(t *testing.T, writer *stubPostWriter, fetcher *stubPostFetcher) (http.Handler, *http.Cookie) {
	t.Helper()
	auth, err := authoauth.NewService(authoauth.Config{
		PublicBaseURL: "https://dev.twisky.app",
		SessionSecret: "test-secret-at-least-32-bytes-long!!",
		StorePath:     filepath.Join(t.TempDir(), "oauth.db"),
		SecureCookies: true,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	t.Cleanup(func() { _ = auth.Close() })

	queries := query.NewDispatcher(
		profile.NewHandler(stubReader{}, nil, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil, nil),
	)
	server := twiskyhttp.NewServer(
		queries,
		command.NewDispatcher(),
		suggestions.NewHandler(stubReader{}, nil, nil),
		"https://dev.twisky.app",
		auth,
		nil,
		nil,
	)
	if writer != nil {
		server = server.WithPostWriter(writer)
	}
	if fetcher != nil {
		server = server.WithPostFetcher(fetcher)
	}

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts:  []session.Account{{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return server.Handler(), recCookie.Result().Cookies()[0]
}

func TestComposeNew_LoggedOutRendersLogin(t *testing.T) {
	t.Parallel()

	handler, _ := newAuthTestServer(t, "https://dev.twisky.app")
	req := httptest.NewRequest(http.MethodGet, "/my/posts/new", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Login with Bluesky") {
		t.Fatalf("body missing login page")
	}
}

func TestComposeNew_LoggedInRendersPage(t *testing.T) {
	t.Parallel()

	handler, cookie := newComposeTestServer(t, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/my/posts/new", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`>New post<`,
		`id="new-post-text"`,
		`action="/my/posts"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %q", body, want)
		}
	}
}

func TestComposeNew_WithParentRendersReply(t *testing.T) {
	t.Parallel()

	parentURI := "at://did:plc:bob/app.bsky.feed.post/parent1"
	fetcher := &stubPostFetcher{posts: []bluesky.Post{{
		URI: parentURI,
		CID: "bafyparent",
		Author: bluesky.Author{
			DID:    "did:plc:bob",
			Handle: "bob.test",
		},
		Record: bluesky.PostRecord{Text: "parent body"},
	}}}
	handler, cookie := newComposeTestServer(t, nil, fetcher)
	req := httptest.NewRequest(http.MethodGet, "/my/posts/new?parent="+url.QueryEscape(parentURI), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`>Reply<`,
		"parent body",
		`name="parent"`,
		parentURI,
		`class="post inset-post"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %q", body, want)
		}
	}
}

func TestCreatePost_EmptyTextReturns400(t *testing.T) {
	t.Parallel()

	writer := &stubPostWriter{}
	handler, cookie := newComposeTestServer(t, writer, nil)
	req := httptest.NewRequest(http.MethodPost, "/my/posts", strings.NewReader("text=+++"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if writer.calls != 0 {
		t.Fatalf("calls = %d, want 0", writer.calls)
	}
	if !strings.Contains(rec.Body.String(), "Post text is required") {
		t.Fatalf("body = %q, want validation message", rec.Body.String())
	}
}

func TestCreatePost_SuccessRedirectsToPost(t *testing.T) {
	t.Parallel()

	writer := &stubPostWriter{}
	handler, cookie := newComposeTestServer(t, writer, nil)
	req := httptest.NewRequest(http.MethodPost, "/my/posts", strings.NewReader("text=hello+world"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303; body=%s", rec.Code, rec.Body.String())
	}
	location := rec.Header().Get("Location")
	if location != "/alice.test/post/newpost1" {
		t.Fatalf("Location = %q, want /alice.test/post/newpost1", location)
	}
	if writer.calls != 1 || writer.text != "hello world" {
		t.Fatalf("CreatePost calls=%d text=%q", writer.calls, writer.text)
	}
	if writer.reply != nil {
		t.Fatalf("reply = %#v, want nil", writer.reply)
	}
}

func TestCreatePost_WithParentForwardsReplyRefs(t *testing.T) {
	t.Parallel()

	parentURI := "at://did:plc:bob/app.bsky.feed.post/parent1"
	rootURI := "at://did:plc:root/app.bsky.feed.post/root1"
	writer := &stubPostWriter{}
	fetcher := &stubPostFetcher{posts: []bluesky.Post{{
		URI: parentURI,
		CID: "bafyparent",
		Author: bluesky.Author{
			DID:    "did:plc:bob",
			Handle: "bob.test",
		},
		Record: bluesky.PostRecord{
			Text: "parent body",
			Reply: &bluesky.RecordReplyRef{
				Root:   bluesky.StrongRef{URI: rootURI, CID: "bafyroot"},
				Parent: bluesky.StrongRef{URI: rootURI, CID: "bafyroot"},
			},
		},
	}}}
	handler, cookie := newComposeTestServer(t, writer, fetcher)
	form := url.Values{
		"text":   {"hello reply"},
		"parent": {parentURI},
	}
	req := httptest.NewRequest(http.MethodPost, "/my/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303; body=%s", rec.Code, rec.Body.String())
	}
	if writer.calls != 1 || writer.text != "hello reply" {
		t.Fatalf("CreatePost calls=%d text=%q", writer.calls, writer.text)
	}
	want := &intent.ReplyTo{
		RootURI:   rootURI,
		RootCID:   "bafyroot",
		ParentURI: parentURI,
		ParentCID: "bafyparent",
	}
	if writer.reply == nil || *writer.reply != *want {
		t.Fatalf("reply = %#v, want %#v", writer.reply, want)
	}
}

func TestCreatePost_InvalidParentDoesNotCreate(t *testing.T) {
	t.Parallel()

	writer := &stubPostWriter{}
	fetcher := &stubPostFetcher{}
	handler, cookie := newComposeTestServer(t, writer, fetcher)
	form := url.Values{
		"text":   {"hello reply"},
		"parent": {"at://did:plc:bob/app.bsky.feed.post/missing"},
	}
	req := httptest.NewRequest(http.MethodPost, "/my/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if writer.calls != 0 {
		t.Fatalf("calls = %d, want 0", writer.calls)
	}
	if !strings.Contains(rec.Body.String(), "Could not load the post to reply to") {
		t.Fatalf("body = %q, want parent error", rec.Body.String())
	}
}

func TestCreatePost_NoSessionReturns401(t *testing.T) {
	t.Parallel()

	handler, _ := newAuthTestServer(t, "https://dev.twisky.app")
	req := httptest.NewRequest(http.MethodPost, "/my/posts", strings.NewReader("text=hello"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestReplyToFromPost_UsesParentAsRoot(t *testing.T) {
	t.Parallel()

	parent := bluesky.Post{
		URI: "at://did:plc:bob/app.bsky.feed.post/parent1",
		CID: "bafyparent",
	}
	got := twiskyhttp.ReplyToFromPost(parent)
	want := intent.ReplyTo{
		RootURI:   parent.URI,
		RootCID:   parent.CID,
		ParentURI: parent.URI,
		ParentCID: parent.CID,
	}
	if got != want {
		t.Fatalf("ReplyToFromPost() = %#v, want %#v", got, want)
	}
}

func TestReplyToFromPost_PreservesThreadRoot(t *testing.T) {
	t.Parallel()

	parent := bluesky.Post{
		URI: "at://did:plc:bob/app.bsky.feed.post/parent1",
		CID: "bafyparent",
		Record: bluesky.PostRecord{
			Reply: &bluesky.RecordReplyRef{
				Root: bluesky.StrongRef{
					URI: "at://did:plc:root/app.bsky.feed.post/root1",
					CID: "bafyroot",
				},
			},
		},
	}
	got := twiskyhttp.ReplyToFromPost(parent)
	if got.RootURI != "at://did:plc:root/app.bsky.feed.post/root1" || got.RootCID != "bafyroot" {
		t.Fatalf("root = %q %q, want thread root", got.RootURI, got.RootCID)
	}
	if got.ParentURI != parent.URI || got.ParentCID != parent.CID {
		t.Fatalf("parent = %q %q, want parent strong ref", got.ParentURI, got.ParentCID)
	}
}
