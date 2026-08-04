package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

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
		t.Fatalf("html = %q, want two mini-profiles", html)
	}
	if strings.Count(html, `class="account-list-item-overlay"`) != 2 {
		t.Fatalf("html = %q, want two row overlays", html)
	}
	if !strings.Contains(html, `href="/one.example"`) {
		t.Fatalf("html = %q, want overlay/profile href for one", html)
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
	if !strings.Contains(html, `data-author-interactive="peekable"`) {
		t.Fatalf("html = %q, want peekable author controls", html)
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
