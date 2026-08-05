package ui

import (
	g "maragu.dev/gomponents"
)

type IconName string

const (
	IconBack     IconName = "back"
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
	IconBrand    IconName = "brand"
)

const iconsSpritePath = "/static/icons/icons.svg"

// iconAtlas maps IconName to the sprite symbol base (icon-<base>-outline/filled).
var iconAtlas = map[IconName]string{
	IconLike:    "heart",
	IconBluesky: "butterfly",
}

// iconStatic maps IconName to a single sprite symbol (id + viewBox).
var iconStatic = map[IconName]struct {
	ID      string
	ViewBox string
}{
	IconBrand: {ID: "icon-brand", ViewBox: "0 0 256 128"},
}

var iconGlyphs = map[IconName]string{
	IconBack:     "⬅️",
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
	IconBrand:    "Twisky",
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

// Icon renders an icon by name. Toggle atlas icons use outline + filled
// layers; static atlas icons use a single symbol; others fall back to glyphs.
func Icon(name IconName) g.Node {
	if base, ok := iconAtlas[name]; ok {
		return svgToggleIcon(base)
	}
	if sym, ok := iconStatic[name]; ok {
		return svgStaticIcon(sym.ID, sym.ViewBox)
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

func svgStaticIcon(symbolID, viewBox string) g.Node {
	return g.El("svg",
		g.Attr("class", "ui-icon"),
		g.Attr("viewBox", viewBox),
		g.Attr("width", "1em"),
		g.Attr("height", "1em"),
		g.Attr("aria-hidden", "true"),
		g.Attr("focusable", "false"),
		g.El("use",
			g.Attr("href", iconsSpritePath+"#"+symbolID),
		),
	)
}
