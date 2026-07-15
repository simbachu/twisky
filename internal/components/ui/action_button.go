package ui

import (
	"fmt"
	"strconv"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ActionButtonConfig describes an icon button with optional engagement count and a11y state.
type ActionButtonConfig struct {
	Icon     IconName
	Label    string
	Count    int
	Disabled bool
	HasPopup bool
}

// PostEngagement builds config for post footer action buttons.
func PostEngagement(icon IconName, label string, count int) ActionButtonConfig {
	return ActionButtonConfig{Icon: icon, Label: label, Count: count}
}

// ActionButton renders a button with an icon, aria-label, and optional count.
// Counts of zero or less are omitted from the label and markup.
//
// TODO: Wire disabled state from engagement rules and auth (e.g. Like when logged out).
// TODO: Add aria-haspopup on Reply, Repost, Share, and More when adding popovers.
func ActionButton(cfg ActionButtonConfig) g.Node {
	return actionButtonNode(cfg, "")
}

func actionButtonNode(cfg ActionButtonConfig, class string) g.Node {
	attrs := []g.Node{
		Icon(cfg.Icon),
		g.Attr("aria-label", actionLabel(cfg.Label, cfg.Count)),
	}
	if class != "" {
		attrs = append([]g.Node{g.Attr("class", class)}, attrs...)
	}
	if cfg.Disabled {
		attrs = append(attrs, g.Attr("disabled", ""), g.Attr("aria-disabled", "true"))
	}
	if cfg.HasPopup {
		attrs = append(attrs, g.Attr("aria-haspopup", "true"))
	}
	if cfg.Count > 0 {
		attrs = append(attrs, Span(
			g.Attr("aria-hidden", "true"),
			g.Text(strconv.Itoa(cfg.Count)),
		))
	}
	return Button(g.Group(attrs))
}

func actionLabel(label string, count int) string {
	if count > 0 {
		return fmt.Sprintf("%s, %d", label, count)
	}
	return label
}
