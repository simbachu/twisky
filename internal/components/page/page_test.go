package page_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
)

func TestPage_ComposesSiteNavAndLoggedInAccountMenu(t *testing.T) {
	t.Parallel()

	current := ui.AuthorInfo{
		Handle:      "alice.test",
		DID:         "did:plc:alice",
		DisplayName: "Alice",
	}
	var buf bytes.Buffer
	if err := page.Page(
		page.PageMeta{Title: "Home", Path: "/"},
		nil,
		ui.AccountMenuView{Enabled: true, Current: &current},
		g.Text("main"),
	).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Site"`) {
		t.Fatalf("html = %q, want Site nav landmark", html)
	}
	if !strings.Contains(html, `class="iface-disclosure account-menu"`) {
		t.Fatalf("html = %q, want disclosure account menu", html)
	}
	if !strings.Contains(html, `id="page-alert"`) {
		t.Fatalf("html = %q, want empty page-alert region", html)
	}
	if !strings.Contains(html, `page-alert.js`) {
		t.Fatalf("html = %q, want page-alert script", html)
	}
	if !strings.Contains(html, `action="/oauth/logout"`) {
		t.Fatalf("html = %q, want logout action in header chrome", html)
	}
	if !strings.Contains(html, "Alice") {
		t.Fatalf("html = %q, want current account name", html)
	}
}

func TestPage_DeclaresMobileViewport(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := page.Page(
		page.PageMeta{Title: "Home", Path: "/"},
		nil,
		ui.AccountMenuView{},
		g.Text("main"),
	).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	want := `name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover"`
	if !strings.Contains(html, want) {
		t.Fatalf("html missing viewport meta %q; got:\n%s", want, html)
	}
}
