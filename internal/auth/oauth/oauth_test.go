package oauth_test

import (
	"net/url"
	"testing"

	"github.com/simbachu/twisky/internal/auth/oauth"
)

func TestLocalhostConfig(t *testing.T) {
	t.Parallel()

	scopes := []string{"atproto", "transition:generic"}
	cfg := oauth.NewConfig("", scopes)

	wantCallback := "http://127.0.0.1:8080/oauth/callback"
	if cfg.CallbackURL != wantCallback {
		t.Fatalf("CallbackURL = %q, want %q", cfg.CallbackURL, wantCallback)
	}

	params := url.Values{}
	params.Set("redirect_uri", wantCallback)
	params.Set("scope", "atproto transition:generic")
	wantClientID := "http://localhost?" + params.Encode()
	if cfg.ClientID != wantClientID {
		t.Fatalf("ClientID = %q, want %q", cfg.ClientID, wantClientID)
	}
}
