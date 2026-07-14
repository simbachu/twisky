package ui

import (
	"fmt"
	"strconv"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ActionButton renders a button with an icon, aria-label, and optional count.
// Counts of zero or less are omitted from the label and markup.
//
// TODO: Replace label+count params with a config struct covering mode,
// disabled state, and aria-haspopup for submenu actions.
// TODO: Wire disabled state from engagement rules and auth (e.g. Like when logged out).
// TODO: Add aria-haspopup on Reply, Repost, Share, and More when adding popovers.
func ActionButton(icon IconName, label string, count int) g.Node {
	return Button(
		Icon(icon),
		g.Attr("aria-label", actionLabel(label, count)),
		g.If(count > 0, Span(
			g.Attr("aria-hidden", "true"),
			g.Text(strconv.Itoa(count)),
		)),
	)
}

func actionLabel(label string, count int) string {
	if count > 0 {
		return fmt.Sprintf("%s, %d", label, count)
	}
	return label
}
