package profile_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/profile"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	profilequery "github.com/simbachu/twisky/internal/query/profile"
)

func TestProfile_RendersPinnedPost(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := profile.Profile(profilequery.ProfileView{
		Handle:      "bsky.app",
		DisplayName: "Bluesky",
		PinnedPostMaybe: &feedquery.PostView{
			ID:                "pinned123",
			AuthorHandle:      "bsky.app",
			AuthorDisplayName: "Bluesky",
			Text:              "pinned hello",
			CreatedAt:         time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Tab: profilequery.TabPosts,
	}, time.Date(2026, 1, 15, 13, 0, 0, 0, time.UTC)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`class="profile-pinned"`,
		`class="profile-pinned-label"`,
		`class="feed-item"`,
		`Pinned`,
		`class="post inset-post"`,
		`pinned hello`,
		`href="/bsky.app/post/pinned123"`,
		`aria-label="View post"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestProfile_OmitsPinnedPostWhenNil(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := profile.Profile(profilequery.ProfileView{
		Handle:      "bsky.app",
		DisplayName: "Bluesky",
		Tab:         profilequery.TabPosts,
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `class="profile-pinned"`) {
		t.Fatalf("html = %q, want no profile-pinned section", html)
	}
}
