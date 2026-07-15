package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestSegmentedGroup_RendersIfaceSegmented(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.SegmentedGroup("Engagement actions",
		ui.PostEngagement(ui.IconReply, "Reply", 2),
		ui.PostEngagement(ui.IconLike, "Like", 5),
	).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="iface-segmented"`) {
		t.Fatalf("html = %q, want iface-segmented class", html)
	}
	if !strings.Contains(html, `aria-label="Engagement actions"`) {
		t.Fatalf("html = %q, want group aria-label", html)
	}
	if !strings.Contains(html, `role="group"`) {
		t.Fatalf("html = %q, want role=group", html)
	}
	if strings.Count(html, "<button") != 2 {
		t.Fatalf("html = %q, want two buttons", html)
	}
}

func TestPillButton_RendersIfacePill(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.PillButton(ui.ActionButtonConfig{
		Icon:  ui.IconFollow,
		Label: "Follow Dev User",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="iface-pill"`) {
		t.Fatalf("html = %q, want iface-pill class", html)
	}
	if !strings.Contains(html, `aria-label="Follow Dev User"`) {
		t.Fatalf("html = %q, want aria-label", html)
	}
}

func TestSearchBar_RendersJoinedControl(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.SearchBar().Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`class="iface-joined"`,
		`role="search"`,
		`method="get"`,
		`action="/search"`,
		`type="search"`,
		`name="q"`,
		`type="submit"`,
		`aria-label="Search"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}
