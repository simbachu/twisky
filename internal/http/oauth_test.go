package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	indigooauth "github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/auth/session"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/command"
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
		profile.NewHandler(stubReader{}, nil, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil, nil),
	)
	handler := twiskyhttp.NewServer(queries, command.NewDispatcher(), suggestions.NewHandler(stubReader{}, nil, nil), baseURL, auth, nil, nil).Handler()
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
	if meta["client_id"] != auth.Config.ClientID {
		t.Fatalf("client_id = %v, want %q", meta["client_id"], auth.Config.ClientID)
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
	if !strings.Contains(body, `method="post"`) {
		t.Fatalf("body missing post method: %s", body)
	}
	if !strings.Contains(body, "Login with Bluesky") {
		t.Fatalf("body missing Login with Bluesky: %s", body)
	}
	if strings.Contains(body, `aria-label="Site"`) {
		t.Fatalf("login page should not include site nav: %s", body)
	}
}

func TestOAuthLogin_DisabledWithoutSecret(t *testing.T) {
	t.Parallel()

	queries := query.NewDispatcher(
		profile.NewHandler(stubReader{}, nil, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil, nil),
	)
	handler := twiskyhttp.NewServer(queries, command.NewDispatcher(), suggestions.NewHandler(stubReader{profile: &bluesky.Profile{Handle: "x"}}, nil, nil), "https://dev.twisky.app", nil, nil, nil).Handler()
	req := httptest.NewRequest(http.MethodGet, "/oauth/client-metadata.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestAccountMenuView_ShowsHandleWhenCookied(t *testing.T) {
	t.Parallel()

	queries := query.NewDispatcher(
		profile.NewHandler(stubReader{
			profile: &bluesky.Profile{Handle: "bsky.app", DID: "did:plc:bsky", DisplayName: "Bluesky"},
			feed:    &bluesky.AuthorFeedResponse{},
		}, nil, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil, nil),
	)
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

	handler := twiskyhttp.NewServer(queries, command.NewDispatcher(), suggestions.NewHandler(stubReader{}, nil, nil), "https://dev.twisky.app", auth, nil, nil).Handler()

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"},
		},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/bsky.app", nil)
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

type stubOAuthApp struct {
	startURL       string
	startErr       error
	callbackData   *indigooauth.ClientSessionData
	callbackErr    error
	logoutErr      error
	resumeErr      error
	logoutCalls    int
	resumeCalls    int
	startCalls     int
	callbackCalls  int
	lastLogoutDID  string
	lastLogoutSess string
	lastStartID    string
}

func (s *stubOAuthApp) StartAuthFlow(_ context.Context, identifier string) (string, error) {
	s.startCalls++
	s.lastStartID = identifier
	return s.startURL, s.startErr
}

func (s *stubOAuthApp) ProcessCallback(_ context.Context, _ url.Values) (*indigooauth.ClientSessionData, error) {
	s.callbackCalls++
	return s.callbackData, s.callbackErr
}

func (s *stubOAuthApp) Logout(_ context.Context, did syntax.DID, sessionID string) error {
	s.logoutCalls++
	s.lastLogoutDID = did.String()
	s.lastLogoutSess = sessionID
	return s.logoutErr
}

func (s *stubOAuthApp) ResumeSession(_ context.Context, did syntax.DID, _ string) (*indigooauth.ClientSession, error) {
	s.resumeCalls++
	if s.resumeErr != nil {
		return nil, s.resumeErr
	}
	return &indigooauth.ClientSession{
		Data: &indigooauth.ClientSessionData{
			AccountDID: did,
			HostURL:    "https://pds.example",
		},
		Config: &indigooauth.ClientConfig{},
	}, nil
}

func newAuthTestServerWithApp(t *testing.T, baseURL string, app authoauth.App) (http.Handler, *authoauth.Service, *twiskyhttp.Server) {
	t.Helper()
	_, auth := newAuthTestServer(t, baseURL)
	auth.App = app
	queries := query.NewDispatcher(
		profile.NewHandler(stubReader{}, nil, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil, nil),
	)
	server := twiskyhttp.NewServer(queries, command.NewDispatcher(), suggestions.NewHandler(stubReader{}, nil, nil), baseURL, auth, nil, nil)
	return server.Handler(), auth, server
}

func TestOAuthLogin_POSTEmptyUsernameRendersError(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{}
	handler, _, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	req := httptest.NewRequest(http.MethodPost, "/oauth/login", strings.NewReader("username="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if app.startCalls != 0 {
		t.Fatalf("StartAuthFlow calls = %d, want 0", app.startCalls)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Handle or DID is required.") {
		t.Fatalf("body missing required error: %s", body)
	}
	if !strings.Contains(body, `action="/oauth/login"`) {
		t.Fatalf("body missing login form: %s", body)
	}
}

func TestOAuthLogin_POSTStartsAuthFlowRedirect(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{startURL: "https://pds.example/authorize?x=1"}
	handler, _, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	req := httptest.NewRequest(http.MethodPost, "/oauth/login", strings.NewReader("username=alice.test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != app.startURL {
		t.Fatalf("Location = %q, want %q", got, app.startURL)
	}
	if app.startCalls != 1 || app.lastStartID != "alice.test" {
		t.Fatalf("StartAuthFlow calls=%d id=%q", app.startCalls, app.lastStartID)
	}
}

func TestOAuthLogin_POSTStartAuthFlowErrorRendersMessage(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{startErr: errors.New("unknown handle")}
	handler, _, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	req := httptest.NewRequest(http.MethodPost, "/oauth/login", strings.NewReader("username=missing.test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Login failed: unknown handle") {
		t.Fatalf("body missing login failure: %s", rec.Body.String())
	}
}

func TestOAuthCallback_SuccessSetsCookieAndRedirects(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{
		callbackData: &indigooauth.ClientSessionData{
			AccountDID: syntax.DID("did:plc:alice"),
			SessionID:  "sess-1",
		},
	}
	handler, _, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=abc&state=xyz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	// Directory lookup is best-effort; without a live directory, handle is empty.
	// Successful login always lands on home.
	if got := rec.Header().Get("Location"); got != "/" {
		t.Fatalf("Location = %q, want /", got)
	}
	if app.callbackCalls != 1 {
		t.Fatalf("ProcessCallback calls = %d, want 1", app.callbackCalls)
	}
	var sessionCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == session.CookieName {
			sessionCookie = c
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatalf("expected session cookie, got %#v", rec.Result().Cookies())
	}
}

func TestOAuthCallback_AddsSecondAccountToExistingCookie(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{
		callbackData: &indigooauth.ClientSessionData{
			AccountDID: syntax.DID("did:plc:bob"),
			SessionID:  "sess-bob",
		},
	}
	handler, auth, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "sess-alice", Handle: "alice.test"},
		},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=abc&state=xyz", nil)
	req.AddCookie(recCookie.Result().Cookies()[0])
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}

	var saved *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == session.CookieName {
			saved = c
		}
	}
	if saved == nil || saved.Value == "" {
		t.Fatalf("expected session cookie, got %#v", rec.Result().Cookies())
	}

	state, err := auth.Jar.Load(cookieRequest(saved))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if state.ActiveDID != "did:plc:bob" {
		t.Fatalf("ActiveDID = %q, want did:plc:bob", state.ActiveDID)
	}
	if len(state.Accounts) != 2 {
		t.Fatalf("len(Accounts) = %d, want 2", len(state.Accounts))
	}
	byDID := map[string]session.Account{}
	for _, account := range state.Accounts {
		byDID[account.DID] = account
	}
	if byDID["did:plc:alice"].SessionID != "sess-alice" {
		t.Fatalf("alice = %#v, want preserved session", byDID["did:plc:alice"])
	}
	if byDID["did:plc:bob"].SessionID != "sess-bob" {
		t.Fatalf("bob = %#v, want new session", byDID["did:plc:bob"])
	}
}

func TestOAuthCallback_ProcessError(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{callbackErr: errors.New("bad state")}
	handler, _, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=abc", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "OAuth callback failed: bad state") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestOAuthLogout_RevokeFailureStillClearsCookie(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{logoutErr: errors.New("revoke failed")}
	handler, auth, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

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
	if app.logoutCalls != 1 || app.lastLogoutDID != "did:plc:alice" || app.lastLogoutSess != "s1" {
		t.Fatalf("Logout calls=%d did=%q sess=%q", app.logoutCalls, app.lastLogoutDID, app.lastLogoutSess)
	}
	cleared := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == session.CookieName && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatalf("expected cleared session cookie after revoke failure, got %#v", rec.Result().Cookies())
	}
}

func TestOAuthLogout_MultiAccountKeepsRemainingCookie(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{}
	handler, auth, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"},
			{DID: "did:plc:bob", SessionID: "s2", Handle: "bob.test"},
		},
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

	var saved *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == session.CookieName {
			saved = c
		}
	}
	if saved == nil || saved.MaxAge < 0 || saved.Value == "" {
		t.Fatalf("expected updated session cookie for remaining account, got %#v", rec.Result().Cookies())
	}

	state, err := auth.Jar.Load(cookieRequest(saved))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if state.ActiveDID != "did:plc:bob" {
		t.Fatalf("ActiveDID = %q, want did:plc:bob", state.ActiveDID)
	}
	if len(state.Accounts) != 1 || state.Accounts[0].DID != "did:plc:bob" {
		t.Fatalf("Accounts = %#v", state.Accounts)
	}
}

func cookieRequest(c *http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(c)
	return req
}

func TestResumeActiveSession_OK(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{}
	_, auth, server := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts:  []session.Account{{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	account, err := server.ResumeActiveSession(cookieRequest(recCookie.Result().Cookies()[0]))
	if err != nil {
		t.Fatalf("ResumeActiveSession: %v", err)
	}
	if account.DID != "did:plc:alice" || account.SessionID != "s1" {
		t.Fatalf("account = %#v", account)
	}
	if app.resumeCalls != 1 {
		t.Fatalf("ResumeSession calls = %d, want 1", app.resumeCalls)
	}
}

func TestResumeActiveSession_MissingCookie(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{}
	_, _, server := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	_, err := server.ResumeActiveSession(httptest.NewRequest(http.MethodGet, "/", nil))
	if !errors.Is(err, session.ErrMissingCookie) {
		t.Fatalf("err = %v, want ErrMissingCookie", err)
	}
	if app.resumeCalls != 0 {
		t.Fatalf("ResumeSession calls = %d, want 0", app.resumeCalls)
	}
}

func TestResumeActiveSession_ResumeError(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{resumeErr: errors.New("expired")}
	_, auth, server := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts:  []session.Account{{DID: "did:plc:alice", SessionID: "s1"}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	_, err := server.ResumeActiveSession(cookieRequest(recCookie.Result().Cookies()[0]))
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("err = %v, want expired", err)
	}
}

func TestOAuthSwitch_SetsActiveDID(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{}
	handler, auth, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"},
			{DID: "did:plc:bob", SessionID: "s2", Handle: "bob.test"},
		},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/oauth/switch", strings.NewReader("did=did:plc:bob"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "dev.twisky.app"
	req.Header.Set("Referer", "https://dev.twisky.app/bsky.app")
	req.AddCookie(recCookie.Result().Cookies()[0])
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/bsky.app" {
		t.Fatalf("Location = %q, want /bsky.app", got)
	}

	var saved *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == session.CookieName {
			saved = c
		}
	}
	if saved == nil || saved.Value == "" {
		t.Fatalf("expected updated session cookie, got %#v", rec.Result().Cookies())
	}
	state, err := auth.Jar.Load(cookieRequest(saved))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if state.ActiveDID != "did:plc:bob" {
		t.Fatalf("ActiveDID = %q, want did:plc:bob", state.ActiveDID)
	}
	if len(state.Accounts) != 2 {
		t.Fatalf("len(Accounts) = %d, want 2", len(state.Accounts))
	}
	if app.resumeCalls != 1 {
		t.Fatalf("ResumeSession calls = %d, want 1", app.resumeCalls)
	}
}

func TestOAuthSwitch_UnknownDID(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{}
	handler, auth, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts:  []session.Account{{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/oauth/switch", strings.NewReader("did=did:plc:missing"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(recCookie.Result().Cookies()[0])
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if app.resumeCalls != 0 {
		t.Fatalf("ResumeSession calls = %d, want 0", app.resumeCalls)
	}
}

func TestOAuthSwitch_MissingCookie(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{}
	handler, _, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	req := httptest.NewRequest(http.MethodPost, "/oauth/switch", strings.NewReader("did=did:plc:alice"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestOAuthSwitch_GETNotAllowed(t *testing.T) {
	t.Parallel()

	handler, _ := newAuthTestServer(t, "https://dev.twisky.app")
	req := httptest.NewRequest(http.MethodGet, "/oauth/switch", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestOAuthSwitch_ResumeFailureRemovesAccount(t *testing.T) {
	t.Parallel()

	app := &stubOAuthApp{resumeErr: errors.New("expired")}
	handler, auth, _ := newAuthTestServerWithApp(t, "https://dev.twisky.app", app)

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"},
			{DID: "did:plc:bob", SessionID: "s2", Handle: "bob.test"},
		},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/oauth/switch", strings.NewReader("did=did:plc:bob"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(recCookie.Result().Cookies()[0])
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/" {
		t.Fatalf("Location = %q, want /", got)
	}

	var saved *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == session.CookieName {
			saved = c
		}
	}
	if saved == nil || saved.Value == "" {
		t.Fatalf("expected updated session cookie, got %#v", rec.Result().Cookies())
	}
	state, err := auth.Jar.Load(cookieRequest(saved))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if state.ActiveDID != "did:plc:alice" {
		t.Fatalf("ActiveDID = %q, want did:plc:alice", state.ActiveDID)
	}
	if len(state.Accounts) != 1 || state.Accounts[0].DID != "did:plc:alice" {
		t.Fatalf("Accounts = %#v, want alice only", state.Accounts)
	}
}

func TestOAuthLogin_GETAddAnotherAccountWhenCookied(t *testing.T) {
	t.Parallel()

	handler, auth := newAuthTestServer(t, "https://dev.twisky.app")
	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts:  []session.Account{{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	req.AddCookie(recCookie.Result().Cookies()[0])
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Add another account") {
		t.Fatalf("body missing add-account copy: %s", body)
	}
	if !strings.Contains(body, `action="/oauth/login"`) {
		t.Fatalf("body missing login form: %s", body)
	}
}
