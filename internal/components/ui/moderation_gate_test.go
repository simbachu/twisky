package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func TestModerationGate_RendersLockedDivWhenNoOverride(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ModerationGate(ui.ModerationGateConfig{
		Message:    "Adult content",
		NoOverride: true,
		Content:    P(g.Text("hidden")),
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `<div class="post-moderation-gate"`) {
		t.Fatalf("html = %q, want locked gate div", html)
	}
	if strings.Contains(html, "<details") {
		t.Fatalf("html = %q, want no disclosure when locked", html)
	}
	if !strings.Contains(html, "Adult content") {
		t.Fatalf("html = %q, want cover message", html)
	}
	if strings.Contains(html, "hidden") {
		t.Fatalf("html = %q, want content omitted when locked", html)
	}
}

func TestModerationGate_RendersDisclosureWhenRevealable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ModerationGate(ui.ModerationGateConfig{
		Message:     "Suggestive content",
		RevealLabel: "Show media",
		Content:     Figure(g.Attr("class", "post-images")),
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="iface-disclosure post-moderation-gate"`) {
		t.Fatalf("html = %q, want disclosure composed with post-moderation-gate", html)
	}
	if !strings.Contains(html, "Show media") {
		t.Fatalf("html = %q, want reveal label", html)
	}
	if !strings.Contains(html, "Suggestive content") {
		t.Fatalf("html = %q, want cover message", html)
	}
	if !strings.Contains(html, `class="post-images"`) {
		t.Fatalf("html = %q, want gated content", html)
	}
	if strings.Contains(html, "aria-expanded") || strings.Contains(html, "aria-haspopup") {
		t.Fatalf("html = %q, want no popover ARIA", html)
	}
}

func TestModerationGate_DefaultsEmptyMessage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ModerationGate(ui.ModerationGateConfig{
		RevealLabel: "Show anyway",
		Content:     P(g.Text("body")),
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "Content warning") {
		t.Fatalf("html = %q, want default content warning", html)
	}
}
