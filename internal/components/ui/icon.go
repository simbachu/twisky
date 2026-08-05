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
	if name == IconBrand {
		return svgBrandIcon()
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

const brandGradientID = "twisky-brand-fill"
const brandMaskID = "twisky-brand-fill-mask"

func svgBrandIcon() g.Node {
	sym := iconStatic[IconBrand]
	return g.El("svg",
		g.Attr("class", "ui-icon ui-icon--brand"),
		g.Attr("viewBox", sym.ViewBox),
		g.Attr("width", "1em"),
		g.Attr("height", "1em"),
		g.Attr("aria-hidden", "true"),
		g.Attr("focusable", "false"),
		g.El("defs",
			g.El("linearGradient",
				g.Attr("id", brandGradientID),
				g.Attr("gradientUnits", "userSpaceOnUse"),
				g.Attr("x1", "0"),
				g.Attr("y1", "0"),
				g.Attr("x2", "90"),
				g.Attr("y2", "128"),
				g.El("stop", g.Attr("offset", "0%"), g.Attr("stop-color", "var(--brand-twilight-1)")),
				g.El("stop", g.Attr("offset", "32%"), g.Attr("stop-color", "var(--brand-twilight-2)")),
				g.El("stop", g.Attr("offset", "68%"), g.Attr("stop-color", "var(--brand-twilight-3)")),
				g.El("stop", g.Attr("offset", "100%"), g.Attr("stop-color", "var(--brand-twilight-4)")),
			),
			g.El("mask",
				g.Attr("id", brandMaskID),
				g.Attr("maskUnits", "userSpaceOnUse"),
				g.Attr("x", "0"),
				g.Attr("y", "0"),
				g.Attr("width", "256"),
				g.Attr("height", "128"),
				g.El("use",
					g.Attr("href", iconsSpritePath+"#"+sym.ID),
					g.Attr("fill", "#fff"),
				),
			),
		),
		g.El("rect",
			g.Attr("width", "256"),
			g.Attr("height", "128"),
			g.Attr("fill", "url(#"+brandGradientID+")"),
			g.Attr("mask", "url(#"+brandMaskID+")"),
		),
	)
}
