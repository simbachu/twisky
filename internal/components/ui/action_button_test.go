package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestActionButton_OmitsCountSpanWhenZeroAndNoCountID(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconLike, "Like", 0)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, "<span") {
		t.Fatalf("html = %q, want no span for a zero, non-pollable count", html)
	}
	if !strings.Contains(html, `aria-label="Like"`) {
		t.Fatalf("html = %q, want plain aria-label", html)
	}
}

func TestActionButton_RendersFuzzyCountWhenNonPollable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconLike, "Like", 5)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Like, 5"`) {
		t.Fatalf("html = %q, want count in aria-label", html)
	}
	if !strings.Contains(html, ">5<") {
		t.Fatalf("html = %q, want fuzzy count text", html)
	}
	if !strings.Contains(html, `class="fuzzy-number"`) {
		t.Fatalf("html = %q, want fuzzy-number class", html)
	}
}

func TestActionButton_RendersFuzzyAbbreviationWhenNonPollable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconLike, "Like", 10_000)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Like, 10K"`) {
		t.Fatalf("html = %q, want fuzzy count in aria-label", html)
	}
	if !strings.Contains(html, ">10K<") {
		t.Fatalf("html = %q, want fuzzy count text", html)
	}
}

func TestActionButton_RendersPollableSpanEvenAtZero(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagementPollable(ui.IconLike, "Like", 0, "like-count-abc")).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="like-count-abc"`,
		`class="fuzzy-number"`,
		`title="0"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestActionButton_EngagedAddsClass(t *testing.T) {
	t.Parallel()

	cfg := ui.PostEngagement(ui.IconLike, "Like", 1)
	cfg.Engaged = true
	var buf bytes.Buffer
	if err := ui.ActionButton(cfg).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `ui-icon-engaged`) {
		t.Fatalf("html = %q, want ui-icon-engaged", html)
	}
}

func TestActionButton_HxPostAttrs(t *testing.T) {
	t.Parallel()

	cfg := ui.PostEngagement(ui.IconLike, "Like", 1)
	cfg.HxPost = "/action/like"
	cfg.HxVals = `{"uri":"at://x","cid":"c"}`
	var buf bytes.Buffer
	if err := ui.ActionButton(cfg).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	for _, want := range []string{
		`hx-post="/action/like"`,
		`hx-vals="{&#34;uri&#34;:&#34;at://x&#34;,&#34;cid&#34;:&#34;c&#34;}"`,
		`hx-target="this"`,
		`hx-swap="outerHTML"`,
		`hx-disabled-elt="this"`,
		`hx-sync="this:drop"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestActionButton_HrefRendersAnchor(t *testing.T) {
	t.Parallel()

	cfg := ui.PostEngagement(ui.IconReply, "Reply", 2)
	cfg.Href = "/my/posts/new?parent=at://did:plc:example/app.bsky.feed.post/abc123"
	cfg.DataComposeOpen = true
	cfg.AriaHasPopup = "dialog"
	var buf bytes.Buffer
	if err := ui.ActionButton(cfg).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	for _, want := range []string{
		`<a `,
		`href="/my/posts/new?parent=at://did:plc:example/app.bsky.feed.post/abc123"`,
		`data-compose-open`,
		`aria-haspopup="dialog"`,
		`aria-label="Reply, 2"`,
		`icon-comment-outline`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, `<button`) {
		t.Fatalf("html = %q, want anchor not button", html)
	}
	if strings.Contains(html, "🗪") {
		t.Fatalf("html = %q, want SVG comment icon not glyph", html)
	}
}
