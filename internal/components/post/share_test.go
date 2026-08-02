package post

import (
	"bytes"
	"strings"
	"testing"

	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestBlueskyPostPageURL(t *testing.T) {
	t.Parallel()

	got := blueskyPostPageURL("example.com", "abc")
	want := "https://bsky.app/profile/example.com/post/abc"
	if got != want {
		t.Fatalf("blueskyPostPageURL() = %q, want %q", got, want)
	}
}

func TestShareGroup_RendersCopyTargets(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := shareGroup(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "dev.example",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`class="iface-segmented post-share-group"`,
		`aria-label="Share"`,
		`class="post-share-open"`,
		`aria-expanded="false"`,
		`aria-haspopup="menu"`,
		`class="post-share-option"`,
		`data-copy-url="/dev.example/post/abc123"`,
		`data-copy-url="https://bsky.app/profile/dev.example/post/abc123"`,
		`aria-label="Copy Twisky link"`,
		`aria-label="Copy Bluesky link"`,
		`data-copy-feedback="icon"`,
		`class="ui-action ui-action--bluesky"`,
		`href="/static/icons/icons.svg#icon-butterfly-outline"`,
		`href="/static/icons/icons.svg#icon-butterfly-filled"`,
		`>🔗<`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, "🦋") {
		t.Fatalf("html = %q, want butterfly SVG not emoji", html)
	}
}
