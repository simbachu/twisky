package ui

import (
	"fmt"
	"time"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func FormatRelativeTime(createdAt, now time.Time) string {
	createdAt = createdAt.UTC()
	now = now.UTC()
	elapsed := now.Sub(createdAt)

	if elapsed < time.Minute {
		return "now"
	}
	if elapsed < time.Hour {
		return fmt.Sprintf("%dm", int(elapsed.Minutes()))
	}
	if elapsed < 24*time.Hour {
		return fmt.Sprintf("%dh", int(elapsed.Hours()))
	}
	if elapsed < 7*24*time.Hour {
		return fmt.Sprintf("%dd", int(elapsed.Hours()/24))
	}
	return createdAt.Format("Jan 2")
}

func FormatAbsoluteTime(createdAt time.Time) string {
	return createdAt.UTC().Format("Jan 2, 2006, 3:04 PM UTC")
}

func Timestamp(createdAt, now time.Time) g.Node {
	if createdAt.IsZero() {
		return nil
	}
	return Span(
		g.Text(" · "),
		Time(
			g.Attr("datetime", createdAt.UTC().Format(time.RFC3339)),
			g.Attr("title", FormatAbsoluteTime(createdAt)),
			g.Text(FormatRelativeTime(createdAt, now)),
		),
	)
}
