package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
)

func TestEngagementStats_RendersDisclosureComposition(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.EngagementStats(ui.EngagementStatsConfig{
		ID:        "engagement-stats-root",
		StatsHref: "/bsky.app/post/root?counts=1",
		Open:      true,
		Summary:   g.Text("2h"),
		Rows: []ui.EngagementStatRow{
			{Label: "Replies", ID: "reply-stats-root", Count: 3},
			{Label: "Reposts", ID: "repost-stats-root", Count: 4},
			{Label: "Likes", ID: "like-stats-root", Count: 5},
		},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`class="iface-disclosure post-engagement-stats"`,
		`id="engagement-stats-root"`,
		`open=""`,
		`data-stats-href="/bsky.app/post/root?counts=1"`,
		"<summary>2h</summary>",
		"<dt>Replies</dt>",
		`id="reply-stats-root"`,
		"<dt>Reposts</dt>",
		`id="repost-stats-root"`,
		"<dt>Likes</dt>",
		`id="like-stats-root"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, "aria-expanded") || strings.Contains(html, "aria-haspopup") {
		t.Fatalf("html = %q, want no popover ARIA", html)
	}
}
