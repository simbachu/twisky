package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type TabItem struct {
	Label   string
	Href    string
	Current bool
}

// TabNav renders a row of navigation tabs.
func TabNav(label string, tabs []TabItem) g.Node {
	if len(tabs) == 0 {
		return nil
	}

	navAttrs := []g.Node{}
	if label != "" {
		navAttrs = append(navAttrs, g.Attr("aria-label", label))
	}

	return Nav(
		g.Group(navAttrs),
		Ul(
			g.Attr("role", "tablist"),
			g.Group(g.Map(tabs, tabItem)),
		),
	)
}

func tabItem(tab TabItem) g.Node {
	return Li(
		g.Attr("role", "presentation"),
		A(
			g.Attr("role", "tab"),
			g.Attr("href", tab.Href),
			g.Attr("aria-selected", ariaSelected(tab.Current)),
			g.If(tab.Current, g.Attr("aria-current", "page")),
			g.Text(tab.Label),
		),
	)
}

func ariaSelected(current bool) string {
	if current {
		return "true"
	}
	return "false"
}
