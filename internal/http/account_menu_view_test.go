package http

import (
	"testing"

	"github.com/simbachu/twisky/internal/auth/session"
)

func TestAccountMenuViewFromState_MapsCurrentAndAdditional(t *testing.T) {
	t.Parallel()

	view := accountMenuViewFromState(session.State{
		ActiveDID: "did:plc:alice",
		Accounts: []session.Account{
			{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"},
			{DID: "did:plc:bob", SessionID: "s2", Handle: "bob.test"},
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
	if len(view.Additional) != 1 {
		t.Fatalf("Additional len = %d, want 1", len(view.Additional))
	}
	if view.Additional[0].Handle != "bob.test" || view.Additional[0].DID != "did:plc:bob" {
		t.Fatalf("Additional[0] = %+v, want bob", view.Additional[0])
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
