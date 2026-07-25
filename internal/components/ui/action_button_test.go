package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestActionButton_OmitsCountSpanWhenZeroAndNoCountID(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconLike, "Like", 0)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, "<span") {
		t.Fatalf("html = %q, want no span for a zero, non-pollable count", html)
	}
	if !strings.Contains(html, `aria-label="Like"`) {
		t.Fatalf("html = %q, want plain aria-label", html)
	}
}

func TestActionButton_RendersFuzzyCountWhenNonPollable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconLike, "Like", 5)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Like, 5"`) {
		t.Fatalf("html = %q, want count in aria-label", html)
	}
	if !strings.Contains(html, ">5<") {
		t.Fatalf("html = %q, want fuzzy count text", html)
	}
	if !strings.Contains(html, `class="fuzzy-number"`) {
		t.Fatalf("html = %q, want fuzzy-number class", html)
	}
}

func TestActionButton_RendersFuzzyAbbreviationWhenNonPollable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagement(ui.IconLike, "Like", 10_000)).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Like, 10K"`) {
		t.Fatalf("html = %q, want fuzzy count in aria-label", html)
	}
	if !strings.Contains(html, ">10K<") {
		t.Fatalf("html = %q, want fuzzy count text", html)
	}
}

func TestActionButton_RendersPollableSpanEvenAtZero(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagementPollable(ui.IconLike, "Like", 0, "like-count-abc")).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="like-count-abc"`,
		`class="fuzzy-number"`,
		`title="0"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestActionButton_RendersPollableSpanWithFuzzyFormatting(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionButton(ui.PostEngagementPollable(ui.IconLike, "Like", 15_000, "like-count-abc")).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="like-count-abc"`,
		`title="15000"`,
		">15K<",
		`aria-label="Like, 15K"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}
