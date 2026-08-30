package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/auth/session"
	"github.com/simbachu/twisky/internal/command"
	twiskyhttp "github.com/simbachu/twisky/internal/http"
	"github.com/simbachu/twisky/internal/query"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/suggestions"
	"github.com/simbachu/twisky/internal/query/tag"
)

type stubPostWriter struct {
	calls     int
	text      string
	recordURI string
	err       error
}

func (s *stubPostWriter) CreatePost(_ context.Context, text string) (string, error) {
	s.calls++
	s.text = text
	if s.recordURI == "" {
		s.recordURI = "at://did:plc:alice/app.bsky.feed.post/newpost1"
	}
	return s.recordURI, s.err
}

func newComposeTestServer(t *testing.T, writer *stubPostWriter) (http.Handler, *http.Cookie) {
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

	handler, cookie := newComposeTestServer(t, nil)
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

func TestCreatePost_EmptyTextReturns400(t *testing.T) {
	t.Parallel()

	writer := &stubPostWriter{}
	handler, cookie := newComposeTestServer(t, writer)
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
	handler, cookie := newComposeTestServer(t, writer)
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
