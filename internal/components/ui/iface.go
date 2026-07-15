package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SegmentedGroup renders multiple action buttons in a shared pill shell.
func SegmentedGroup(label string, buttons ...ActionButtonConfig) g.Node {
	items := make([]g.Node, len(buttons))
	for i, cfg := range buttons {
		items[i] = Li(ActionButton(cfg))
	}
	return Ul(
		g.Attr("class", "iface-segmented"),
		g.Attr("aria-label", label),
		g.Attr("role", "group"),
		g.Group(items),
	)
}

// PillButton renders a standalone pill-shaped action button.
func PillButton(cfg ActionButtonConfig) g.Node {
	return actionButtonNode(cfg, "iface-pill")
}

// SearchBar renders a joined search input and submit button.
func SearchBar() g.Node {
	return Form(
		g.Attr("class", "iface-joined"),
		g.Attr("role", "search"),
		g.Attr("method", "get"),
		g.Attr("action", "/search"),
		Input(
			g.Attr("type", "search"),
			g.Attr("name", "q"),
			g.Attr("placeholder", "🔍︎ Search"),
			g.Attr("aria-label", "Search"),
			g.Attr("title", "Search the Bluesky network"),
		),
		Button(
			g.Attr("type", "submit"),
			g.Attr("aria-label", "Search"),
			Icon(IconSearch),
		),
	)
}
