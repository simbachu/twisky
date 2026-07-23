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
	// CountID, when set, forces the count span to always render (even at
	// zero) with this id and fuzzy-number formatting, so it can be targeted
	// by an htmx out-of-band swap when live counts polling is enabled.
	CountID string
}

// PostEngagement builds config for post footer action buttons.
func PostEngagement(icon IconName, label string, count int) ActionButtonConfig {
	return ActionButtonConfig{Icon: icon, Label: label, Count: count}
}

// PostEngagementPollable builds config for a post footer action button whose
// count span carries a stable id and fuzzy-number formatting, for posts with
// live counts polling enabled.
func PostEngagementPollable(icon IconName, label string, count int, id string) ActionButtonConfig {
	return ActionButtonConfig{Icon: icon, Label: label, Count: count, CountID: id}
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
	switch {
	case cfg.CountID != "":
		attrs = append(attrs, FuzzyCountSpan(cfg.CountID, cfg.Count, false))
	case cfg.Count > 0:
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
