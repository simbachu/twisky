package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestActionButton_OmitsCountWhenZero(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconReply, "Replies", 0)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.HasPrefix(html, "<button") {
		t.Fatalf("html = %q, want button element", html)
	}
	if strings.Contains(html, "<li") {
		t.Fatalf("html = %q, want no list item wrapper", html)
	}
	if !strings.Contains(html, `aria-label="Replies"`) {
		t.Fatalf("html = %q, want aria-label", html)
	}
	if strings.Contains(html, ">0<") {
		t.Fatalf("html = %q, want no count when zero", html)
	}
}

func TestActionButton_OmitsCountWhenNegative(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconShare, "Share", -1)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Share"`) {
		t.Fatalf("html = %q, want aria-label", html)
	}
	if strings.Contains(html, ">-1<") {
		t.Fatalf("html = %q, want no count when negative", html)
	}
}

func TestActionButton_RendersCountWhenPositive(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconLike, "Likes", 42)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Likes, 42"`) {
		t.Fatalf("html = %q, want aria-label with count", html)
	}
	if !strings.Contains(html, `<span aria-hidden="true">42</span>`) {
		t.Fatalf("html = %q, want visually hidden count span", html)
	}
}

func TestActionButton_RendersDisabled(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cfg := ui.PostEngagement(ui.IconLike, "Like", 0)
	cfg.Disabled = true
	if err := ui.ActionButton(cfg).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `disabled`) {
		t.Fatalf("html = %q, want disabled attribute", html)
	}
	if !strings.Contains(html, `aria-disabled="true"`) {
		t.Fatalf("html = %q, want aria-disabled", html)
	}
}

func TestActionButton_RendersHasPopup(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cfg := ui.PostEngagement(ui.IconMore, "More options", 0)
	cfg.HasPopup = true
	if err := ui.ActionButton(cfg).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-haspopup="true"`) {
		t.Fatalf("html = %q, want aria-haspopup", html)
	}
}
