package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestComposeModal_WrapsFieldInDialog(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ComposeModal(ui.ComposeFieldConfig{
		TextareaID: "compose-text",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`<dialog`,
		`id="compose-dialog"`,
		`action="/my/posts"`,
		`id="compose-text"`,
		`formmethod="dialog"`,
		`aria-label="Close"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %q", html, want)
		}
	}
}
