package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func TestDisclosure_RendersDetailsWithIfaceClass(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.Disclosure("", g.Text("Toggle"), P(g.Text("Panel"))).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<details") {
		t.Fatalf("html = %q, want details", html)
	}
	if !strings.Contains(html, `class="iface-disclosure"`) {
		t.Fatalf("html = %q, want iface-disclosure class", html)
	}
	if !strings.Contains(html, "<summary>Toggle</summary>") {
		t.Fatalf("html = %q, want summary text", html)
	}
	if !strings.Contains(html, "<p>Panel</p>") {
		t.Fatalf("html = %q, want panel content", html)
	}
}

func TestDisclosure_AppendsExtraClass(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.Disclosure("account-menu", Span(g.Text("Summary")), Ul()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="iface-disclosure account-menu"`) {
		t.Fatalf("html = %q, want iface-disclosure account-menu", html)
	}
	if strings.Contains(html, "aria-expanded") || strings.Contains(html, "aria-haspopup") {
		t.Fatalf("html = %q, want no popover ARIA on disclosure", html)
	}
}

func TestDisclosureWith_RendersOptionalAttrs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.DisclosureWith(ui.DisclosureConfig{
		ExtraClass: "post-engagement-stats",
		Summary:    g.Text("Summary"),
		Content:    []g.Node{P(g.Text("Panel"))},
		ID:         "engagement-stats-abc",
		Open:       true,
		Attrs:      []g.Node{g.Attr("data-stats-href", "/x?counts=1")},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="iface-disclosure post-engagement-stats"`) {
		t.Fatalf("html = %q, want iface-disclosure post-engagement-stats", html)
	}
	if !strings.Contains(html, `id="engagement-stats-abc"`) {
		t.Fatalf("html = %q, want id", html)
	}
	if !strings.Contains(html, `open=""`) {
		t.Fatalf("html = %q, want open", html)
	}
	if !strings.Contains(html, `data-stats-href="/x?counts=1"`) {
		t.Fatalf("html = %q, want data-stats-href", html)
	}
	if !strings.Contains(html, "<summary>Summary</summary>") {
		t.Fatalf("html = %q, want summary", html)
	}
	if !strings.Contains(html, "<p>Panel</p>") {
		t.Fatalf("html = %q, want panel", html)
	}
}
