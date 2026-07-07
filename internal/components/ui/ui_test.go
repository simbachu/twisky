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
	if err := ui.ActionButton(ui.IconReply, "Replies", 0).Render(&buf); err != nil {
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
	if err := ui.ActionButton(ui.IconShare, "Share", -1).Render(&buf); err != nil {
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
	if err := ui.ActionButton(ui.IconLike, "Likes", 42).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Likes"`) {
		t.Fatalf("html = %q, want aria-label", html)
	}
	if !strings.Contains(html, "👍42") {
		t.Fatalf("html = %q, want count 42 after icon", html)
	}
}
