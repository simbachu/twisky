package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/simbachu/twisky/internal/auth/session"
)

func TestCookie_RoundTripMultiAccount(t *testing.T) {
	t.Parallel()

	jar := session.NewJar([]byte("test-secret-at-least-32-bytes-long!!"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	state := session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "sess-a", Handle: "alice.bsky.social"},
			{DID: "did:plc:bob", SessionID: "sess-b", Handle: "bob.bsky.social"},
		},
	}
	if err := jar.Save(rec, state); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	req.AddCookie(cookies[0])

	got, err := jar.Load(req)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.ActiveDID != state.ActiveDID {
		t.Fatalf("ActiveDID = %q, want %q", got.ActiveDID, state.ActiveDID)
	}
	if len(got.Accounts) != 2 {
		t.Fatalf("len(Accounts) = %d, want 2", len(got.Accounts))
	}
	if got.Accounts[1].Handle != "bob.bsky.social" {
		t.Fatalf("Accounts[1].Handle = %q", got.Accounts[1].Handle)
	}
}

func TestCookie_TamperedRejected(t *testing.T) {
	t.Parallel()

	jar := session.NewJar([]byte("test-secret-at-least-32-bytes-long!!"))
	rec := httptest.NewRecorder()
	if err := jar.Save(rec, session.State{
		ActiveDID: "did:plc:alice",
		Accounts:  []session.Account{{DID: "did:plc:alice", SessionID: "s"}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	cookie := rec.Result().Cookies()[0]
	cookie.Value = cookie.Value + "tamper"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	if _, err := jar.Load(req); err == nil {
		t.Fatal("Load tampered cookie: want error")
	}
}

func TestState_AddAccountReplacesSameDID(t *testing.T) {
	t.Parallel()

	state := session.State{}
	state = state.AddAccount(session.Account{DID: "did:plc:alice", SessionID: "old", Handle: "alice.test"})
	state = state.AddAccount(session.Account{DID: "did:plc:alice", SessionID: "new", Handle: "alice.test"})
	state = state.AddAccount(session.Account{DID: "did:plc:bob", SessionID: "b", Handle: "bob.test"})

	if len(state.Accounts) != 2 {
		t.Fatalf("len(Accounts) = %d, want 2", len(state.Accounts))
	}
	if state.ActiveDID != "did:plc:bob" {
		t.Fatalf("ActiveDID = %q, want did:plc:bob", state.ActiveDID)
	}
	if state.Accounts[0].SessionID != "new" {
		t.Fatalf("alice SessionID = %q, want new", state.Accounts[0].SessionID)
	}
}

func TestState_RemoveActivePromotesNext(t *testing.T) {
	t.Parallel()

	state := session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "a"},
			{DID: "did:plc:bob", SessionID: "b"},
		},
	}
	state = state.RemoveAccount("did:plc:alice")
	if state.ActiveDID != "did:plc:bob" {
		t.Fatalf("ActiveDID = %q, want did:plc:bob", state.ActiveDID)
	}
	if len(state.Accounts) != 1 {
		t.Fatalf("len(Accounts) = %d, want 1", len(state.Accounts))
	}
}

func TestState_ActiveAccount(t *testing.T) {
	t.Parallel()

	state := session.State{
		ActiveDID: "did:plc:bob",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "a"},
			{DID: "did:plc:bob", SessionID: "b", Handle: "bob.test"},
		},
	}
	account, ok := state.ActiveAccount()
	if !ok {
		t.Fatal("ActiveAccount: want ok")
	}
	if account.Handle != "bob.test" {
		t.Fatalf("Handle = %q", account.Handle)
	}
}

func TestJar_Clear(t *testing.T) {
	t.Parallel()

	jar := session.NewJar([]byte("test-secret-at-least-32-bytes-long!!"))
	rec := httptest.NewRecorder()
	jar.Clear(rec)
	cookie := rec.Result().Cookies()[0]
	if cookie.MaxAge != -1 {
		t.Fatalf("MaxAge = %d, want -1", cookie.MaxAge)
	}
	if cookie.Value != "" {
		t.Fatalf("Value = %q, want empty", cookie.Value)
	}
}
