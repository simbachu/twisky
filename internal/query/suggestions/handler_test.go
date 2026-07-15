package suggestions_test

import (
	"context"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/query/suggestions"
)

type stubReader struct {
	profiles []bluesky.Profile
	err      error
}

func (s stubReader) GetProfiles(context.Context, []string) ([]bluesky.Profile, error) {
	return s.profiles, s.err
}

func TestHandler_SuggestedAccounts_PreservesHandleOrder(t *testing.T) {
	t.Parallel()

	handler := suggestions.NewHandler(stubReader{
		profiles: []bluesky.Profile{
			{Handle: "bsky.app", DisplayName: "Bluesky", Avatar: "https://example.com/bluesky.jpg"},
			{Handle: "simbachu.com", DisplayName: "ſpectral", Avatar: "https://example.com/simbachu.jpg"},
		},
	}, []string{"simbachu.com", "bsky.app"})

	accounts := handler.SuggestedAccounts(context.Background())
	if len(accounts) != 2 {
		t.Fatalf("len(accounts) = %d, want 2", len(accounts))
	}
	if accounts[0].Handle != "simbachu.com" {
		t.Fatalf("accounts[0].Handle = %q, want simbachu.com", accounts[0].Handle)
	}
	if accounts[0].DisplayName != "ſpectral" {
		t.Fatalf("accounts[0].DisplayName = %q, want ſpectral", accounts[0].DisplayName)
	}
	if accounts[1].Handle != "bsky.app" {
		t.Fatalf("accounts[1].Handle = %q, want bsky.app", accounts[1].Handle)
	}
}

func TestHandler_SuggestedAccounts_OmitsMissingProfiles(t *testing.T) {
	t.Parallel()

	handler := suggestions.NewHandler(stubReader{
		profiles: []bluesky.Profile{
			{Handle: "bsky.app", DisplayName: "Bluesky"},
		},
	}, []string{"simbachu.com", "bsky.app"})

	accounts := handler.SuggestedAccounts(context.Background())
	if len(accounts) != 1 {
		t.Fatalf("len(accounts) = %d, want 1", len(accounts))
	}
	if accounts[0].Handle != "bsky.app" {
		t.Fatalf("accounts[0].Handle = %q, want bsky.app", accounts[0].Handle)
	}
}

func TestHandler_SuggestedAccounts_ReturnsNilOnError(t *testing.T) {
	t.Parallel()

	handler := suggestions.NewHandler(stubReader{err: context.Canceled}, nil)
	if accounts := handler.SuggestedAccounts(context.Background()); accounts != nil {
		t.Fatalf("accounts = %v, want nil on upstream error", accounts)
	}
}
