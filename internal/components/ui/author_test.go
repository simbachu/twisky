package ui_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestAuthorLink_ShowsNameAndHandle(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.AuthorLink(ui.AuthorInfo{
		Handle:      "dev.example",
		DisplayName: "Dev User",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="author-name"`) {
		t.Fatalf("html = %q, want author-name", html)
	}
	if !strings.Contains(html, "Dev User") {
		t.Fatalf("html = %q, want display name", html)
	}
	if !strings.Contains(html, "@dev.example") {
		t.Fatalf("html = %q, want handle", html)
	}
}

func TestAuthorLink_OmitsNameWhenSameAsHandle(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.AuthorLink(ui.AuthorInfo{
		Handle:      "dev.example",
		DisplayName: "dev.example",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `class="author-name"`) {
		t.Fatalf("html = %q, want no author-name when display equals handle", html)
	}
	if !strings.Contains(html, "@dev.example") {
		t.Fatalf("html = %q, want handle only", html)
	}
}

func TestTimestamp_RendersRelativeWithAbsoluteTitle(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 12, 18, 43, 28, 0, time.UTC)
	now := time.Date(2026, 3, 12, 20, 43, 28, 0, time.UTC)

	var buf bytes.Buffer
	if err := ui.Timestamp(createdAt, now).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `datetime="2026-03-12T18:43:28Z"`) {
		t.Fatalf("html = %q, want datetime attribute", html)
	}
	if !strings.Contains(html, `title="Mar 12, 2026, 6:43 PM UTC"`) {
		t.Fatalf("html = %q, want absolute title", html)
	}
	if !strings.Contains(html, ">2h<") {
		t.Fatalf("html = %q, want relative visible text", html)
	}
}

func TestTimestamp_OmitsWhenZero(t *testing.T) {
	t.Parallel()

	if ui.Timestamp(time.Time{}, time.Now()) != nil {
		t.Fatal("Timestamp() = node, want nil for zero CreatedAt")
	}
}
