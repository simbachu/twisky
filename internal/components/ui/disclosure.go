package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// DisclosureConfig configures optional details attributes beyond class/summary/content.
type DisclosureConfig struct {
	ExtraClass string
	Summary    g.Node
	Content    []g.Node
	ID         string
	Open       bool
	Attrs      []g.Node
}

// Disclosure renders a native details/summary control.
// extraClass is appended to iface-disclosure when non-empty (e.g. "account-menu").
func Disclosure(extraClass string, summary g.Node, content ...g.Node) g.Node {
	return DisclosureWith(DisclosureConfig{
		ExtraClass: extraClass,
		Summary:    summary,
		Content:    content,
	})
}

// DisclosureWith renders a native details/summary control with optional attrs.
func DisclosureWith(cfg DisclosureConfig) g.Node {
	class := "iface-disclosure"
	if cfg.ExtraClass != "" {
		class += " " + cfg.ExtraClass
	}
	attrs := []g.Node{g.Attr("class", class)}
	if cfg.ID != "" {
		attrs = append(attrs, g.Attr("id", cfg.ID))
	}
	if cfg.Open {
		attrs = append(attrs, g.Attr("open", ""))
	}
	attrs = append(attrs, cfg.Attrs...)
	attrs = append(attrs, Summary(cfg.Summary), g.Group(cfg.Content))
	return Details(attrs...)
}
