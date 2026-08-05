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

func TestSiteNav_PrimarySecondaryAndLabels(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.SiteNav("/").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Count(html, `data-nav="primary"`) != 3 {
		t.Fatalf("html = %q, want three primary nav items", html)
	}
	if strings.Count(html, `data-nav="secondary"`) != 1 {
		t.Fatalf("html = %q, want one secondary nav item", html)
	}
	if !strings.Contains(html, `href="/settings"`) {
		t.Fatalf("html = %q, want settings link", html)
	}
	if strings.Count(html, `aria-hidden="true"`) != 4 {
		t.Fatalf("html = %q, want four aria-hidden icon spans", html)
	}
	for _, label := range []string{"Home", "Explore", "Notifications", "Settings"} {
		if !strings.Contains(html, "<span>"+label+"</span>") {
			t.Fatalf("html = %q, want label span %s", html, label)
		}
	}
}
