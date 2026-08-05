package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SegmentedShell renders a shared pill shell for grouped controls that share semantics.
// extraClass is appended to iface-segmented when non-empty (e.g. "post-share-group").
func SegmentedShell(label, extraClass string, items ...g.Node) g.Node {
	class := "iface-segmented"
	if extraClass != "" {
		class += " " + extraClass
	}
	return Menu(
		g.Attr("class", class),
		g.Attr("aria-label", label),
		g.Attr("role", "group"),
		g.Group(items),
	)
}

// SegmentedGroup renders multiple action buttons in a shared pill shell.
func SegmentedGroup(label string, buttons ...ActionButtonConfig) g.Node {
	items := make([]g.Node, len(buttons))
	for i, cfg := range buttons {
		items[i] = Li(ActionButton(cfg))
	}
	return SegmentedShell(label, "", items...)
}

// SegmentedRadioOption describes one radio in a SegmentedRadioGroup.
type SegmentedRadioOption struct {
	Value string
	Icon  string
	Label string
}

// SegmentedRadioGroup renders radio options in a shared pill shell.
// Input ids are name+"-"+Value. current selects the checked option; empty means none checked.
func SegmentedRadioGroup(name, ariaLabel, current string, options []SegmentedRadioOption) g.Node {
	items := make([]g.Node, len(options))
	for i, opt := range options {
		inputID := name + "-" + opt.Value
		items[i] = Li(
			Label(
				g.Attr("for", inputID),
				g.Attr("title", opt.Label),
				Input(
					g.Attr("type", "radio"),
					g.Attr("id", inputID),
					g.Attr("name", name),
					g.Attr("value", opt.Value),
					g.If(current == opt.Value, g.Attr("checked", "")),
				),
				Span(g.Text(opt.Icon)),
			),
		)
	}
	return SegmentedShell(ariaLabel, "", items...)
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

// BackButton renders a pill link that history.backs when the referrer is
// same-origin; otherwise it navigates to fallbackHref (typically "/").
func BackButton(fallbackHref string) g.Node {
	if fallbackHref == "" {
		fallbackHref = "/"
	}
	return A(
		g.Attr("class", "iface-pill"),
		g.Attr("href", fallbackHref),
		g.Attr("data-nav-back", ""),
		g.Attr("aria-label", "Back"),
		Icon(IconBack),
	)
}
