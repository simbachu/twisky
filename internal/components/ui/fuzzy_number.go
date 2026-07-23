package ui

import (
	"fmt"
	"strconv"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const (
	fuzzyNumberKThreshold = 10_000
	fuzzyNumberMThreshold = 1_000_000
)

// FormatFuzzyNumber abbreviates large counts: exact below 10K, then K (thousands), then M (millions).
func FormatFuzzyNumber(n int) string {
	switch {
	case n >= fuzzyNumberMThreshold:
		return fmt.Sprintf("%dM", n/fuzzyNumberMThreshold)
	case n >= fuzzyNumberKThreshold:
		return fmt.Sprintf("%dK", n/1_000)
	default:
		return strconv.Itoa(n)
	}
}

// FuzzyNumber renders an abbreviated count, with the exact value in title.
func FuzzyNumber(n int) g.Node {
	return Span(
		g.Attr("class", "fuzzy-number"),
		g.Attr("title", strconv.Itoa(n)),
		g.Text(FormatFuzzyNumber(n)),
	)
}

// FuzzyCountSpan renders an abbreviated, aria-hidden count with a stable id so
// it can be targeted by an htmx out-of-band swap. When oob is true, the span
// carries hx-swap-oob so it replaces the element with the same id in place.
func FuzzyCountSpan(id string, n int, oob bool) g.Node {
	attrs := []g.Node{
		g.Attr("id", id),
		g.Attr("class", "fuzzy-number"),
		g.Attr("aria-hidden", "true"),
		g.Attr("title", strconv.Itoa(n)),
	}
	if oob {
		attrs = append(attrs, g.Attr("hx-swap-oob", "true"))
	}
	attrs = append(attrs, g.Text(FormatFuzzyNumber(n)))
	return Span(attrs...)
}
