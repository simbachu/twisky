package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Emoji now but svg later
const (
	IconHome          = "🏠"
	IconExplore       = "🔍"
	IconNotifications = "🔔"
	IconSettings      = "⚙️"
)

const (
	navPrimary   = "primary"
	navSecondary = "secondary"
)

func siteNavItem(tier, icon, label, href string, attrs ...g.Node) g.Node {
	return Li(
		g.Attr("data-nav", tier),
		A(
			g.Attr("href", href),
			g.Group(attrs),
			Span(g.Attr("aria-hidden", "true"), g.Text(icon)),
			Span(g.Text(label)),
		),
	)
}

func siteNavSeparator() g.Node {
	return Li(g.Attr("role", "separator"))
}

func siteNavCurrent(currentPath, href string) g.Node {
	return g.If(currentPath == href, g.Attr("aria-current", "page"))
}

// SiteNav is the primary site navigation landmark (destinations added later).
func SiteNav(currentPath string) g.Node {
	return Nav(g.Attr("aria-label", "Site"),
		Ul(
			siteNavItem(navPrimary, IconHome, "Home", "/", siteNavCurrent(currentPath, "/")),
			siteNavItem(navPrimary, IconExplore, "Explore", "/explore", siteNavCurrent(currentPath, "/explore")),
			siteNavItem(navPrimary, IconNotifications, "Notifications", "/notifications", siteNavCurrent(currentPath, "/notifications")),
			siteNavItem(navSecondary, IconSettings, "Settings", "/settings", siteNavCurrent(currentPath, "/settings")),
		),
	)
}
