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

func siteNavItem(icon, label, href string, attrs ...g.Node) g.Node {
	return Li(A(g.Attr("href", href), g.Group(attrs), g.Text(icon), g.Text(label)))
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
			siteNavItem(IconHome, "Home", "/", siteNavCurrent(currentPath, "/")),
			siteNavItem(IconExplore, "Explore", "/explore", siteNavCurrent(currentPath, "/explore")),
			siteNavItem(IconNotifications, "Notifications", "/notifications", siteNavCurrent(currentPath, "/notifications")),
			siteNavItem(IconSettings, "Settings", "/settings", siteNavCurrent(currentPath, "/settings")),
		),
	)
}
