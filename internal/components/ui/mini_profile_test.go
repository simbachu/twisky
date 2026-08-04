package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestMiniProfile_PeekableRendersDiscreteLinks(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.MiniProfile(ui.AuthorInfo{
		Handle:      "dev.example",
		DisplayName: "Dev User",
		Avatar:      "https://example.com/avatar.jpg",
	}, ui.AuthorPeekable).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.HasPrefix(html, `<span class="mini-profile"`) {
		t.Fatalf("html = %q, want mini-profile span chrome", html)
	}
	if !strings.Contains(html, `class="profile-icon"`) {
		t.Fatalf("html = %q, want profile-icon", html)
	}
	if !strings.Contains(html, `src="https://example.com/avatar.jpg"`) {
		t.Fatalf("html = %q, want avatar src", html)
	}
	if !strings.Contains(html, `class="author-name"`) {
		t.Fatalf("html = %q, want author-name", html)
	}
	if !strings.Contains(html, "Dev User") {
		t.Fatalf("html = %q, want display name", html)
	}
	if !strings.Contains(html, "@dev.example") {
		t.Fatalf("html = %q, want handle", html)
	}
	if strings.Count(html, `href="/dev.example"`) < 3 {
		t.Fatalf("html = %q, want avatar, name, and handle hrefs", html)
	}
	if strings.Count(html, `data-author-interactive="peekable"`) < 3 {
		t.Fatalf("html = %q, want peekable markers on avatar/name/handle", html)
	}
	if strings.Contains(html, "<button") {
		t.Fatalf("html = %q, want no button inside mini-profile", html)
	}
}

func TestMiniProfile_StaticOmitsProfileLinks(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.MiniProfile(ui.AuthorInfo{
		Handle:      "alice.test",
		DisplayName: "Alice",
		Avatar:      "https://example.com/a.jpg",
	}, ui.AuthorStatic).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `href=`) {
		t.Fatalf("html = %q, want no profile hrefs", html)
	}
	if strings.Contains(html, `data-author-interactive=`) {
		t.Fatalf("html = %q, want no interactive markers", html)
	}
	if !strings.Contains(html, "Alice") {
		t.Fatalf("html = %q, want display name", html)
	}
}
