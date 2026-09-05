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

func TestShareGroup_RendersCopyURLs(t *testing.T) {
	t.Parallel()

	html := renderShareGroup(t)

	if !strings.Contains(html, `data-copy-url="/dev.example/post/abc123"`) {
		t.Fatalf("html = %q, want Twisky copy URL", html)
	}
	if !strings.Contains(html, `data-copy-url="https://bsky.app/profile/dev.example/post/abc123"`) {
		t.Fatalf("html = %q, want Bluesky copy URL", html)
	}
	if !strings.Contains(html, `aria-label="Copy Twisky link"`) || !strings.Contains(html, `aria-label="Copy Bluesky link"`) {
		t.Fatalf("html = %q, want copy aria labels", html)
	}
}

func TestShareGroup_RendersMenuChrome(t *testing.T) {
	t.Parallel()

	html := renderShareGroup(t)

	for _, want := range []string{
		`class="iface-segmented post-share-group"`,
		`class="post-share-open"`,
		`aria-expanded="false"`,
		`aria-haspopup="menu"`,
		`data-copy-feedback="icon"`,
		`href="/static/icons/icons.svg#icon-share-outline"`,
		`href="/static/icons/icons.svg#icon-butterfly-outline"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, "↗") {
		t.Fatalf("html = %q, want share SVG not glyph", html)
	}
	if strings.Contains(html, "🦋") {
		t.Fatalf("html = %q, want butterfly SVG not emoji", html)
	}
}

func renderShareGroup(t *testing.T) string {
	t.Helper()
	var buf bytes.Buffer
	if err := shareGroup(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "dev.example",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	return buf.String()
}
