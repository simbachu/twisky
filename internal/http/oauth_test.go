package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/auth/session"
	"github.com/simbachu/twisky/internal/bluesky"
	twiskyhttp "github.com/simbachu/twisky/internal/http"
	"github.com/simbachu/twisky/internal/query"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/suggestions"
	"github.com/simbachu/twisky/internal/query/tag"
)

func newAuthTestServer(t *testing.T, baseURL string) (http.Handler, *authoauth.Service) {
	t.Helper()
	auth, err := authoauth.NewService(authoauth.Config{
		PublicBaseURL: baseURL,
		SessionSecret: "test-secret-at-least-32-bytes-long!!",
		StorePath:     filepath.Join(t.TempDir(), "oauth.db"),
		SecureCookies: strings.HasPrefix(baseURL, "https://"),
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	t.Cleanup(func() { _ = auth.Close() })

	queries := query.NewDispatcher(
		profile.NewHandler(stubReader{}, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil),
	)
	handler := twiskyhttp.NewServer(queries, suggestions.NewHandler(stubReader{}, nil), baseURL, auth).Handler()
	return handler, auth
}

func TestClientMetadata_PublicJSON(t *testing.T) {
	t.Parallel()

	handler, auth := newAuthTestServer(t, "https://dev.twisky.app")
	req := httptest.NewRequest(http.MethodGet, "/oauth/client-metadata.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("Content-Type = %q", ct)
	}

	var meta map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil {
		t.Fatalf("json: %v", err)
	}
	if meta["client_id"] != auth.App.Config.ClientID {
		t.Fatalf("client_id = %v, want %q", meta["client_id"], auth.App.Config.ClientID)
	}
	redirects, ok := meta["redirect_uris"].([]any)
	if !ok || len(redirects) == 0 || redirects[0] != "https://dev.twisky.app/oauth/callback" {
		t.Fatalf("redirect_uris = %#v", meta["redirect_uris"])
	}
	if meta["client_name"] != "Twisky" {
		t.Fatalf("client_name = %v", meta["client_name"])
	}
}

func TestOAuthLogin_GETRendersForm(t *testing.T) {
	t.Parallel()

	handler, _ := newAuthTestServer(t, "https://dev.twisky.app")
	req := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `action="/oauth/login"`) {
		t.Fatalf("body missing login form: %s", body)
	}
	if !strings.Contains(body, `name="username"`) {
		t.Fatalf("body missing username field: %s", body)
	}
	if !strings.Contains(body, "Log in") {
		t.Fatalf("body missing Log in chrome: %s", body)
	}
}

func TestOAuthLogin_DisabledWithoutSecret(t *testing.T) {
	t.Parallel()

	queries := query.NewDispatcher(
		profile.NewHandler(stubReader{}, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil),
	)
	handler := twiskyhttp.NewServer(queries, suggestions.NewHandler(stubReader{profile: &bluesky.Profile{Handle: "x"}}, nil), "https://dev.twisky.app", nil).Handler()
	req := httptest.NewRequest(http.MethodGet, "/oauth/client-metadata.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestAuthChrome_ShowsHandleWhenCookied(t *testing.T) {
	t.Parallel()

	handler, auth := newAuthTestServer(t, "https://dev.twisky.app")
	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"},
		},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	req.AddCookie(recCookie.Result().Cookies()[0])
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "alice.test") {
		t.Fatalf("body missing handle: %s", body)
	}
	if !strings.Contains(body, `action="/oauth/logout"`) {
		t.Fatalf("body missing logout: %s", body)
	}
}

func TestOAuthLogout_ClearsCookie(t *testing.T) {
	t.Parallel()

	handler, auth := newAuthTestServer(t, "https://dev.twisky.app")
	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts:  []session.Account{{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/oauth/logout", nil)
	req.AddCookie(recCookie.Result().Cookies()[0])
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	cleared := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == session.CookieName && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatalf("expected cleared session cookie, got %#v", rec.Result().Cookies())
	}
}
