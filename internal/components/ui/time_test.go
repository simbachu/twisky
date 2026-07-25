package ui_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestFormatRelativeTimeLong_JustNow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 25, 14, 56, 0, 0, time.UTC)
	createdAt := now.Add(-30 * time.Second)

	if got := ui.FormatRelativeTimeLong(createdAt, now); got != "just now" {
		t.Fatalf("FormatRelativeTimeLong() = %q, want just now", got)
	}
}

func TestFormatRelativeTimeLong_Minutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 25, 14, 56, 0, 0, time.UTC)

	if got := ui.FormatRelativeTimeLong(now.Add(-1*time.Minute), now); got != "1 minute ago" {
		t.Fatalf("FormatRelativeTimeLong() = %q, want 1 minute ago", got)
	}
	if got := ui.FormatRelativeTimeLong(now.Add(-21*time.Minute), now); got != "21 minutes ago" {
		t.Fatalf("FormatRelativeTimeLong() = %q, want 21 minutes ago", got)
	}
}

func TestFormatRelativeTimeLong_Hours(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 25, 14, 56, 0, 0, time.UTC)

	if got := ui.FormatRelativeTimeLong(now.Add(-time.Hour), now); got != "1 hour ago" {
		t.Fatalf("FormatRelativeTimeLong() = %q, want 1 hour ago", got)
	}
	if got := ui.FormatRelativeTimeLong(now.Add(-2*time.Hour), now); got != "2 hours ago" {
		t.Fatalf("FormatRelativeTimeLong() = %q, want 2 hours ago", got)
	}
}

func TestFormatRelativeTimeLong_DaysAndWeeks(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 25, 14, 56, 0, 0, time.UTC)

	if got := ui.FormatRelativeTimeLong(now.Add(-3*24*time.Hour), now); got != "3 days ago" {
		t.Fatalf("FormatRelativeTimeLong() = %q, want 3 days ago", got)
	}
	if got := ui.FormatRelativeTimeLong(now.Add(-14*24*time.Hour), now); got != "2 weeks ago" {
		t.Fatalf("FormatRelativeTimeLong() = %q, want 2 weeks ago", got)
	}
}

func TestFormatPostPageTime(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 25, 14, 56, 0, 0, time.UTC)
	createdAt := now.Add(-2 * time.Hour)

	want := "2026-07-25 12:56 UTC (2 hours ago)"
	if got := ui.FormatPostPageTime(createdAt, now); got != want {
		t.Fatalf("FormatPostPageTime() = %q, want %q", got, want)
	}
}

func TestPostPageTimestamp_RendersTimeElement(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 25, 14, 56, 0, 0, time.UTC)
	createdAt := now.Add(-2 * time.Hour)

	var buf bytes.Buffer
	if err := ui.PostPageTimestamp(createdAt, now).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`class="post-page-timestamp"`,
		`datetime="2026-07-25T12:56:00Z"`,
		"2026-07-25 12:56 UTC (2 hours ago)",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %q", html, want)
		}
	}
}

func TestFormatRelativeTime_UnderOneMinute(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 20, 43, 28, 0, time.UTC)
	createdAt := now.Add(-59 * time.Second)

	if got := ui.FormatRelativeTime(createdAt, now); got != "now" {
		t.Fatalf("FormatRelativeTime() = %q, want now", got)
	}
}

func TestFormatRelativeTime_Minutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 20, 43, 28, 0, time.UTC)
	createdAt := now.Add(-21 * time.Minute)

	if got := ui.FormatRelativeTime(createdAt, now); got != "21m" {
		t.Fatalf("FormatRelativeTime() = %q, want 21m", got)
	}
}

func TestFormatRelativeTime_Hours(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 20, 43, 28, 0, time.UTC)
	createdAt := now.Add(-90 * time.Minute)

	if got := ui.FormatRelativeTime(createdAt, now); got != "1h" {
		t.Fatalf("FormatRelativeTime() = %q, want 1h", got)
	}
}

func TestFormatRelativeTime_Days(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 20, 43, 28, 0, time.UTC)
	createdAt := now.Add(-3 * 24 * time.Hour)

	if got := ui.FormatRelativeTime(createdAt, now); got != "3d" {
		t.Fatalf("FormatRelativeTime() = %q, want 3d", got)
	}
}

func TestFormatRelativeTime_OlderThanSevenDays(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 20, 43, 28, 0, time.UTC)
	createdAt := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	if got := ui.FormatRelativeTime(createdAt, now); got != "Feb 1" {
		t.Fatalf("FormatRelativeTime() = %q, want Feb 1", got)
	}
}

func TestFormatAbsoluteTime(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 12, 20, 43, 28, 0, time.UTC)

	if got := ui.FormatAbsoluteTime(createdAt); got != "Mar 12, 2026, 8:43 PM UTC" {
		t.Fatalf("FormatAbsoluteTime() = %q, want Mar 12, 2026, 8:43 PM UTC", got)
	}
}
