package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestTabNav_RendersTablistStructure(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.TabNav("Profile", []ui.TabItem{
		{Label: "Posts", Href: "/bsky.app", Current: true},
		{Label: "Media", Href: "/bsky.app/media", Current: false},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<nav") {
		t.Fatalf("html = %q, want nav element", html)
	}
	if !strings.Contains(html, `role="tablist"`) {
		t.Fatalf("html = %q, want tablist role", html)
	}
	if !strings.Contains(html, `role="tab"`) {
		t.Fatalf("html = %q, want tab role", html)
	}
	if !strings.Contains(html, `role="presentation"`) {
		t.Fatalf("html = %q, want presentation role on list items", html)
	}
}

func TestTabNav_MarksCurrentTab(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.TabNav("Profile", []ui.TabItem{
		{Label: "Posts", Href: "/bsky.app", Current: true},
		{Label: "Media", Href: "/bsky.app/media", Current: false},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	postsTab := `<a role="tab" href="/bsky.app" aria-selected="true" aria-current="page">Posts</a>`
	if !strings.Contains(html, postsTab) {
		t.Fatalf("html = %q, want current Posts tab markup", html)
	}
	mediaTab := `<a role="tab" href="/bsky.app/media" aria-selected="false">Media</a>`
	if !strings.Contains(html, mediaTab) {
		t.Fatalf("html = %q, want inactive Media tab markup", html)
	}
}

func TestTabNav_OmitsNavWhenEmpty(t *testing.T) {
	t.Parallel()

	if ui.TabNav("Profile", nil) != nil {
		t.Fatal("TabNav() = node, want nil for empty tabs")
	}
}

func TestTabNav_SetsAriaLabel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.TabNav("Profile", []ui.TabItem{
		{Label: "Posts", Href: "/bsky.app", Current: true},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Profile"`) {
		t.Fatalf("html = %q, want aria-label on nav", html)
	}
}
