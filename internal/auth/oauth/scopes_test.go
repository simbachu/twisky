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
		"repo:app.bsky.feed.like?action=create",
		"repo:app.bsky.feed.like?action=delete",
		"repo:app.bsky.feed.repost?action=create",
		"repo:app.bsky.feed.repost?action=delete",
		"repo:app.bsky.feed.post?action=create",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("DefaultScopes = %q, missing %q", joined, want)
		}
	}
	for _, invalid := range []string{
		"transition:generic",
		"authManagePrefs",
		"action=manage",
	} {
		if strings.Contains(joined, invalid) {
			t.Fatalf("DefaultScopes = %q, must not include %q", joined, invalid)
		}
	}
}
