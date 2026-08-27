package oauth_test

import (
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/auth/oauth"
)

func TestDefaultScopes_IncludeTimelineAndPrefs(t *testing.T) {
	t.Parallel()

	joined := strings.Join(oauth.DefaultScopes, " ")
	for _, want := range []string{
		"atproto",
		"include:app.bsky.authViewAll?aud=did:web:api.bsky.app%23bsky_appview",
		"include:app.bsky.authManagePrefs",
		"repo:app.bsky.feed.like?action=manage",
		"repo:app.bsky.feed.repost?action=manage",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("DefaultScopes = %q, missing %q", joined, want)
		}
	}
	if strings.Contains(joined, "transition:generic") {
		t.Fatalf("DefaultScopes still includes transition:generic: %q", joined)
	}
}
