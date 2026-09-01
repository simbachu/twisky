package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestIcon_ThreadRendersStaticSprite(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.Icon(ui.IconThread).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `href="/static/icons/icons.svg#icon-thread-outline"`) {
		t.Fatalf("html = %q, want thread outline sprite", html)
	}
	if strings.Contains(html, "🧵") {
		t.Fatalf("html = %q, want SVG sprite not emoji fallback", html)
	}
}
