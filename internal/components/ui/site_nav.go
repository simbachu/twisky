package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SiteNav is the primary site navigation landmark (destinations added later).
func SiteNav() g.Node {
	return Nav(g.Attr("aria-label", "Site"))
}
