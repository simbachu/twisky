package ui

import (
	"strconv"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ActionButton renders a button with an icon, aria-label, and optional count.
// Counts of zero or less are omitted.
func ActionButton(icon IconName, label string, count int) g.Node {
	return Button(
		Icon(icon),
		g.Attr("aria-label", label),
		g.If(count > 0, g.Text(strconv.Itoa(count))),
	)
}
