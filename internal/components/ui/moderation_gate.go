package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ModerationGateConfig describes a locked or revealable moderation cover.
type ModerationGateConfig struct {
	Message     string
	RevealLabel string
	NoOverride  bool
	Content     g.Node
}

// ModerationGate renders a locked cover or a Disclosure that reveals Content.
func ModerationGate(cfg ModerationGateConfig) g.Node {
	message := cfg.Message
	if message == "" {
		message = "Content warning"
	}
	cover := P(g.Text(message))
	if cfg.NoOverride {
		return Div(g.Attr("class", "post-moderation-gate"), cover)
	}
	return Disclosure(
		"post-moderation-gate",
		g.Group{cover, g.Text(cfg.RevealLabel)},
		cfg.Content,
	)
}
