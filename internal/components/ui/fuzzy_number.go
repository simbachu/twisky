package ui

import (
	"fmt"
	"strconv"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FormatFuzzyNumber abbreviates large counts with K or M suffixes.
// Currently threshold is 10K.
func FormatFuzzyNumber(n int) string {
	suffix := ""
	value := n
	if n >= 1_000_000 {
		value = n / 1_000_000
		suffix = "M"
	} else if n >= 1_0000 {
		value = n / 1_0000
		suffix = "K"
	}
	return fmt.Sprintf("%d%s", value, suffix)
}

// FuzzyNumber renders an abbreviated count, with the exact value in title.
func FuzzyNumber(n int) g.Node {
	return Span(
		g.Attr("class", "fuzzy-number"),
		g.Attr("title", strconv.Itoa(n)),
		g.Text(FormatFuzzyNumber(n)),
	)
}
