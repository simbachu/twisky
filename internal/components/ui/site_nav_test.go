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
	if err := ui.SiteNav("/").Render(&buf); err != nil {
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

func TestSiteNav_MarksCurrentPage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.SiteNav("/explore").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	current := `<a href="/explore" aria-current="page">`
	if !strings.Contains(html, current) {
		t.Fatalf("html = %q, want current Explore link markup", html)
	}
	if strings.Count(html, `aria-current="page"`) != 1 {
		t.Fatalf("html = %q, want exactly one aria-current", html)
	}
	if strings.Contains(html, `<a href="/" aria-current="page">`) {
		t.Fatalf("html = %q, Home must not be current on /explore", html)
	}
}
