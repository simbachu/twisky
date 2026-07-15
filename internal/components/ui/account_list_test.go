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

func TestAccountList_RendersAllItems(t *testing.T) {
	t.Parallel()

	accounts := []ui.AuthorInfo{
		{Handle: "one.example", DisplayName: "One", Avatar: "https://example.com/one.jpg"},
		{Handle: "two.example", DisplayName: "Two", Avatar: "https://example.com/two.jpg"},
	}

	var buf bytes.Buffer
	if err := ui.AccountList(accounts).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="account-list"`) {
		t.Fatalf("html = %q, want account-list class", html)
	}
	if strings.Count(html, `class="mini-profile"`) != 2 {
		t.Fatalf("html = %q, want two mini-profile links", html)
	}
	if strings.Count(html, `class="iface-pill"`) != 2 {
		t.Fatalf("html = %q, want two iface-pill buttons", html)
	}
	if !strings.Contains(html, `aria-label="Follow One"`) {
		t.Fatalf("html = %q, want Follow One aria-label", html)
	}
	if !strings.Contains(html, `aria-label="Follow Two"`) {
		t.Fatalf("html = %q, want Follow Two aria-label", html)
	}
	if !strings.Contains(html, "One") || !strings.Contains(html, "Two") {
		t.Fatalf("html = %q, want both display names", html)
	}
}

func TestAccountList_OmitsWhenEmpty(t *testing.T) {
	t.Parallel()

	if ui.AccountList(nil) != nil {
		t.Fatal("AccountList(nil) = node, want nil")
	}
	if ui.AccountList([]ui.AuthorInfo{}) != nil {
		t.Fatal("AccountList([]) = node, want nil")
	}
}
