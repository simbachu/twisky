package ui

import (
	"fmt"
	"time"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const PostPageDateLayout = "2006-01-02 15:04" // phpBB equivalent: 'Y-m-d H:i'

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

func FormatRelativeTimeLong(createdAt, now time.Time) string {
	createdAt = createdAt.UTC()
	now = now.UTC()
	elapsed := now.Sub(createdAt)

	if elapsed < time.Minute {
		return "just now"
	}
	if elapsed < time.Hour {
		return pluralAgo(int(elapsed.Minutes()), "minute")
	}
	if elapsed < 24*time.Hour {
		return pluralAgo(int(elapsed.Hours()), "hour")
	}
	days := int(elapsed.Hours() / 24)
	if days < 7 {
		return pluralAgo(days, "day")
	}
	weeks := days / 7
	if weeks < 5 {
		return pluralAgo(weeks, "week")
	}
	months := days / 30
	if months < 12 {
		return pluralAgo(months, "month")
	}
	return pluralAgo(days/365, "year")
}

func pluralAgo(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s ago", unit)
	}
	return fmt.Sprintf("%d %ss ago", n, unit)
}

func FormatPostPageTime(createdAt, now time.Time) string {
	absolute := createdAt.UTC().Format(PostPageDateLayout)
	timezone := "UTC" // TODO: get timezone from user settings
	relative := FormatRelativeTimeLong(createdAt, now)
	return fmt.Sprintf("%s %s (%s)", absolute, timezone, relative)
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

func PostPageTimestamp(createdAt, now time.Time) g.Node {
	if createdAt.IsZero() {
		return nil
	}
	return Time(
		g.Attr("class", "post-page-timestamp"),
		g.Attr("datetime", createdAt.UTC().Format(time.RFC3339)),
		g.Attr("title", FormatAbsoluteTime(createdAt)),
		g.Text(FormatPostPageTime(createdAt, now)),
	)
}
