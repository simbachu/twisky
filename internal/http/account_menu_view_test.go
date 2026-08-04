package http

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/auth/session"
	"github.com/simbachu/twisky/internal/components/ui"
)

func TestAccountMenuViewFromState_MapsCurrentAndAdditional(t *testing.T) {
	t.Parallel()

	view := accountMenuViewFromState(session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{
				DID:         "did:plc:alice",
				SessionID:   "s1",
				Handle:      "alice.test",
				DisplayName: "Alice",
				Avatar:      "https://cdn.example/alice.jpg",
			},
			{
				DID:         "did:plc:bob",
				SessionID:   "s2",
				Handle:      "bob.test",
				DisplayName: "Bob",
				Avatar:      "https://cdn.example/bob.jpg",
			},
		},
	})

	if !view.Enabled {
		t.Fatal("Enabled = false, want true")
	}
	if view.Current == nil {
		t.Fatal("Current = nil, want active account")
	}
	if view.Current.Handle != "alice.test" || view.Current.DID != "did:plc:alice" {
		t.Fatalf("Current = %+v, want alice", view.Current)
	}
	if view.Current.DisplayName != "Alice" || view.Current.Avatar != "https://cdn.example/alice.jpg" {
		t.Fatalf("Current chrome = %+v, want DisplayName/Avatar", view.Current)
	}
	if len(view.Additional) != 1 {
		t.Fatalf("Additional len = %d, want 1", len(view.Additional))
	}
	if view.Additional[0].Handle != "bob.test" || view.Additional[0].DID != "did:plc:bob" {
		t.Fatalf("Additional[0] = %+v, want bob", view.Additional[0])
	}
	if view.Additional[0].DisplayName != "Bob" || view.Additional[0].Avatar != "https://cdn.example/bob.jpg" {
		t.Fatalf("Additional[0] chrome = %+v, want DisplayName/Avatar", view.Additional[0])
	}
}

func TestAccountMenuViewFromState_FallsBackDisplayNameToHandle(t *testing.T) {
	t.Parallel()

	view := accountMenuViewFromState(session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"},
		},
	})
	if view.Current == nil {
		t.Fatal("Current = nil, want active account")
	}
	if view.Current.DisplayName != "alice.test" {
		t.Fatalf("DisplayName = %q, want handle fallback", view.Current.DisplayName)
	}
}

func TestAccountMenuViewFromState_LoggedOutWhenNoActive(t *testing.T) {
	t.Parallel()

	view := accountMenuViewFromState(session.State{})
	if !view.Enabled {
		t.Fatal("Enabled = false, want true")
	}
	if view.Current != nil {
		t.Fatalf("Current = %+v, want nil", view.Current)
	}
	if len(view.Additional) != 0 {
		t.Fatalf("Additional = %v, want empty", view.Additional)
	}
}

func testAuthServer(t *testing.T) *Server {
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
	return &Server{auth: auth}
}

func TestAccountMenuView_SyncsKnownAuthorChrome(t *testing.T) {
	t.Parallel()

	s := testAuthServer(t)
	recCookie := httptest.NewRecorder()
	if err := s.auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"},
		},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(recCookie.Result().Cookies()[0])
	rec := httptest.NewRecorder()

	view := s.accountMenuView(rec, req, ui.AuthorInfo{
		Handle:      "alice.test",
		DisplayName: "Alice",
		DID:         "did:plc:alice",
		Avatar:      "https://cdn.example/alice.jpg",
	})
	if view.Current == nil {
		t.Fatal("Current = nil")
	}
	if view.Current.DisplayName != "Alice" || view.Current.Avatar != "https://cdn.example/alice.jpg" {
		t.Fatalf("Current = %+v, want synced chrome", view.Current)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Set-Cookie count = %d, want 1", len(cookies))
	}
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(cookies[0])
	got, err := s.auth.Jar.Load(req2)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Accounts[0].DisplayName != "Alice" || got.Accounts[0].Avatar != "https://cdn.example/alice.jpg" {
		t.Fatalf("cookie account = %+v, want persisted chrome", got.Accounts[0])
	}
}

func TestAccountMenuView_IgnoresUnrelatedAuthor(t *testing.T) {
	t.Parallel()

	s := testAuthServer(t)
	recCookie := httptest.NewRecorder()
	if err := s.auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"},
		},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(recCookie.Result().Cookies()[0])
	rec := httptest.NewRecorder()

	view := s.accountMenuView(rec, req, ui.AuthorInfo{
		Handle:      "bob.test",
		DisplayName: "Bob",
		DID:         "did:plc:bob",
		Avatar:      "https://cdn.example/bob.jpg",
	})
	if view.Current == nil || view.Current.DisplayName != "alice.test" {
		t.Fatalf("Current = %+v, want handle fallback unchanged", view.Current)
	}
	if strings.Contains(rec.Header().Get("Set-Cookie"), session.CookieName+"=") {
		t.Fatalf("unexpected Set-Cookie: %q", rec.Header().Get("Set-Cookie"))
	}
}

func TestAccountMenuView_NoAuthNoop(t *testing.T) {
	t.Parallel()

	s := &Server{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	view := s.accountMenuView(rec, req, ui.AuthorInfo{DID: "did:plc:alice", DisplayName: "Alice"})
	if view.Enabled {
		t.Fatalf("view = %+v, want disabled", view)
	}
}
