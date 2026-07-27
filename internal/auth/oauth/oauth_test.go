package oauth_test

import (
	"testing"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
)

func TestLocalhostConfig(t *testing.T) {
	t.Parallel()

	cfg := oauth.NewLocalhostConfig("http://127.0.0.1:8080/oauth/callback", []string{"atproto", "transition:generic"})
	if cfg.CallbackURL != "http://127.0.0.1:8080/oauth/callback" {
		t.Fatalf("CallbackURL = %q, want http://127.0.0.1:8080/oauth/callback", cfg.CallbackURL)
	}
	if cfg.ClientID == "" {
		t.Fatal("ClientID is empty")
	}
}
