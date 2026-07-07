package ui_test

import (
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/ui"
)

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
