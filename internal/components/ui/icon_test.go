package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestIcon_LikeRendersHeartAtlas(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.Icon(ui.IconLike).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	for _, want := range []string{
		`class="ui-icon ui-icon--toggle"`,
		`href="/static/icons/icons.svg#icon-heart-outline"`,
		`href="/static/icons/icons.svg#icon-heart-filled"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, "👍") {
		t.Fatalf("html = %q, want SVG atlas not emoji", html)
	}
}

func TestIcon_ReplyRendersCommentSprite(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.Icon(ui.IconReply).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	for _, want := range []string{
		`class="ui-icon"`,
		`href="/static/icons/icons.svg#icon-comment-outline"`,
		`viewBox="0 0 64 64"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, "🗪") {
		t.Fatalf("html = %q, want SVG sprite not glyph", html)
	}
	if strings.Contains(html, "ui-icon--toggle") {
		t.Fatalf("html = %q, want static icon not toggle", html)
	}
}

func TestIcon_UpRendersGlyph(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.Icon(ui.IconUp).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	if html != "↑" {
		t.Fatalf("html = %q, want ↑ glyph", html)
	}
}

func TestIcon_BrandRendersStaticSprite(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.Icon(ui.IconBrand).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	for _, want := range []string{
		`class="ui-icon ui-icon--brand"`,
		`viewBox="0 0 256 128"`,
		`width="2em"`,
		`height="1em"`,
		`gradientUnits="userSpaceOnUse"`,
		`id="twisky-brand-fill"`,
		`id="twisky-brand-fill-mask"`,
		`href="/static/icons/icons.svg#icon-brand"`,
		`fill="#fff"`,
		`fill="url(#twisky-brand-fill)"`,
		`mask="url(#twisky-brand-fill-mask)"`,
		`<rect`,
		`stop-color="var(--brand-twilight-1)"`,
		`aria-hidden="true"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, `width="1em"`) {
		t.Fatalf("html = %q, brand must not use square 1em width", html)
	}
	if strings.Contains(html, "ui-icon--toggle") {
		t.Fatalf("html = %q, want static icon not toggle", html)
	}
}

func TestActionButton_LikeIncludesActionClass(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconLike, "Like", 0)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `class="ui-action ui-action--like"`) {
		t.Fatalf("html = %q, want ui-action--like class", html)
	}
	if !strings.Contains(html, `icon-heart-outline`) {
		t.Fatalf("html = %q, want heart outline symbol", html)
	}
}

func TestActionButton_RepostIncludesActionClass(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconRepost, "Repost", 0)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `class="ui-action ui-action--repost"`) {
		t.Fatalf("html = %q, want ui-action--repost class", html)
	}
}
