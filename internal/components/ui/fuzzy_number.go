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
