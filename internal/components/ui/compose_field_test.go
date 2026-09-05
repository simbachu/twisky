package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestComposeField_RendersPostForm(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ComposeField(ui.ComposeFieldConfig{
		TextareaID: "compose-text",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`<form`,
		`method="post"`,
		`action="/my/posts"`,
		`<textarea`,
		`name="text"`,
		`id="compose-text"`,
		`type="submit"`,
		`>Post<`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %q", html, want)
		}
	}
}

func TestComposeField_PreservesTextAndError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ComposeField(ui.ComposeFieldConfig{
		TextareaID: "new-post-text",
		Text:       "hello world",
		Error:      "Post text is required.",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="new-post-text"`,
		">hello world<",
		`role="alert"`,
		"Post text is required.",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %q", html, want)
		}
	}
}

func TestComposeField_RendersHiddenParent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ComposeField(ui.ComposeFieldConfig{
		TextareaID: "compose-text",
		ParentURI:  "at://did:plc:example/app.bsky.feed.post/parent1",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`type="hidden"`,
		`name="parent"`,
		`value="at://did:plc:example/app.bsky.feed.post/parent1"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %q", html, want)
		}
	}
}
