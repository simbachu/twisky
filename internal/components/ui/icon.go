package ui

import (
	g "maragu.dev/gomponents"
)

type IconName string

const (
	IconReply    IconName = "reply"
	IconRepost   IconName = "repost"
	IconLike     IconName = "like"
	IconShare    IconName = "share"
	IconBookmark IconName = "bookmark"
	IconMore     IconName = "more"
	IconFollow   IconName = "follow"
	IconSearch   IconName = "search"
	IconPlay     IconName = "play"
	IconPause    IconName = "pause"
	IconBluesky  IconName = "bluesky"
)

const iconsSpritePath = "/static/icons/icons.svg"

// iconAtlas maps IconName to the sprite symbol base (icon-<base>-outline/filled).
var iconAtlas = map[IconName]string{
	IconLike:    "heart",
	IconBluesky: "butterfly",
}

var iconGlyphs = map[IconName]string{
	IconReply:    "🗪",
	IconRepost:   "⇄",
	IconLike:     "👍︎",
	IconShare:    "↗",
	IconBookmark: "𖤘",
	IconMore:     "⋯",
	IconFollow:   "Follow",
	IconSearch:   "🔍︎",
	IconPlay:     "▶",
	IconPause:    "⏸",
	IconBluesky:  "🦋",
}

// ActionClass returns a CSS modifier for per-action hover colors, or empty.
func ActionClass(name IconName) string {
	switch name {
	case IconLike:
		return "ui-action ui-action--like"
	case IconBluesky:
		return "ui-action ui-action--bluesky"
	default:
		return ""
	}
}

// Icon renders an icon by name. Atlas icons use the packed SVG sprite
// (outline + filled layers); others fall back to emoji placeholders.
func Icon(name IconName) g.Node {
	if base, ok := iconAtlas[name]; ok {
		return svgToggleIcon(base)
	}
	return g.Text(iconGlyphs[name])
}

func svgToggleIcon(base string) g.Node {
	outlineID := "icon-" + base + "-outline"
	filledID := "icon-" + base + "-filled"
	return g.El("svg",
		g.Attr("class", "ui-icon ui-icon--toggle"),
		g.Attr("viewBox", "0 0 64 64"),
		g.Attr("width", "1em"),
		g.Attr("height", "1em"),
		g.Attr("aria-hidden", "true"),
		g.Attr("focusable", "false"),
		g.El("use",
			g.Attr("class", "ui-icon-outline"),
			g.Attr("href", iconsSpritePath+"#"+outlineID),
		),
		g.El("use",
			g.Attr("class", "ui-icon-filled"),
			g.Attr("href", iconsSpritePath+"#"+filledID),
		),
	)
}
