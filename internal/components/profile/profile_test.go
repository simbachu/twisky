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
	}, time.Date(2026, 1, 15, 13, 0, 0, 0, time.UTC), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`class="profile-pinned"`,
		`class="profile-pinned-label"`,
		`class="clickable-inset"`,
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
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `class="profile-pinned"`) {
		t.Fatalf("html = %q, want no profile-pinned section", html)
	}
}

func TestProfile_RendersSocialMetaTags(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := profile.Profile(profilequery.ProfileView{
		Handle:      "simbachu.com",
		DisplayName: "Spectral",
		Description: "gamer, mother, husband",
		Avatar:      "https://cdn.example/avatar.jpg",
		Followers:   404,
		Following:   237,
		Posts:       663,
		Tab:         profilequery.TabPosts,
	}, time.Now().UTC(), nil, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:title" content="Spectral (@simbachu.com)"`,
		`property="og:description" content="gamer, mother, husband"`,
		`property="og:image" content="https://cdn.example/avatar.jpg"`,
		`property="og:type" content="profile"`,
		`property="og:url" content="https://twisky.test/simbachu.com"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestProfile_UsesStatsWhenDescriptionMissing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := profile.Profile(profilequery.ProfileView{
		Handle:    "bsky.app",
		DisplayName: "Bluesky",
		Followers: 100,
		Following: 50,
		Posts:     25,
		Tab:       profilequery.TabPosts,
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `property="og:description" content="100 followers · 50 following · 25 posts"`) {
		t.Fatalf("html = %q, want stats-based description", html)
	}
}
