package ui

import (
	g "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

// DialogConfig configures a native dialog element.
type DialogConfig struct {
	ID         string
	ExtraClass string
	Content    []g.Node
}

// NativeDialog renders a native <dialog> shell.
func NativeDialog(cfg DialogConfig) g.Node {
	class := "iface-dialog"
	if cfg.ExtraClass != "" {
		class += " " + cfg.ExtraClass
	}
	attrs := []g.Node{g.Attr("class", class)}
	if cfg.ID != "" {
		attrs = append(attrs, g.Attr("id", cfg.ID))
	}
	attrs = append(attrs, g.Group(cfg.Content))
	return html.Dialog(attrs...)
}
