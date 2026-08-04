package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestSiteNav_RendersLandmark(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.SiteNav().Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<nav") {
		t.Fatalf("html = %q, want nav element", html)
	}
	if !strings.Contains(html, `aria-label="Site"`) {
		t.Fatalf("html = %q, want Site aria-label", html)
	}
}
