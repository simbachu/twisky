package ui

import (
	"strconv"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const (
	groupedNumberThreshold = 10_000
	thinSpace              = "\u2009"
)

// FormatGroupedNumber renders an exact integer with thin-space thousands
// separators when n >= 10_000. Below that, returns strconv.Itoa(n).
func FormatGroupedNumber(n int) string {
	if n < groupedNumberThreshold {
		return strconv.Itoa(n)
	}

	s := strconv.Itoa(n)
	var out []byte
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, thinSpace...)
		}
		out = append(out, byte(digit))
	}
	return string(out)
}

// GroupedStatCount renders a grouped exact count in a definition value with a
// stable id so it can be targeted by an htmx out-of-band swap.
func GroupedStatCount(id string, n int, oob bool) g.Node {
	attrs := []g.Node{
		g.Attr("id", id),
		g.Attr("title", strconv.Itoa(n)),
	}
	if oob {
		attrs = append(attrs, g.Attr("hx-swap-oob", "true"))
	}
	attrs = append(attrs, g.Text(FormatGroupedNumber(n)))
	return Dd(attrs...)
}
