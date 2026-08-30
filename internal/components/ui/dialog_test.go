package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
)

func TestDialog_RendersNativeElementWithID(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.NativeDialog(ui.DialogConfig{
		ID:         "compose-dialog",
		ExtraClass: "compose-dialog",
		Content:    []g.Node{g.Text("body")},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		"<dialog",
		`id="compose-dialog"`,
		`class="iface-dialog compose-dialog"`,
		"body",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %q", html, want)
		}
	}
}
