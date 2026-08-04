package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestMiniProfile_RendersAvatarNameAndHandle(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.MiniProfile(ui.AuthorInfo{
		Handle:      "dev.example",
		DisplayName: "Dev User",
		Avatar:      "https://example.com/avatar.jpg",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="mini-profile"`) {
		t.Fatalf("html = %q, want mini-profile class", html)
	}
	if !strings.Contains(html, `href="/dev.example"`) {
		t.Fatalf("html = %q, want profile href", html)
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
	if strings.Contains(html, "<button") {
		t.Fatalf("html = %q, want no button inside profile link", html)
	}
}
