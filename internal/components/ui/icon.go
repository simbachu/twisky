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
)

var iconGlyphs = map[IconName]string{
	IconReply:    "🗪",
	IconRepost:   "⇄",
	IconLike:     "👍︎",
	IconShare:    "↗",
	IconBookmark: "𖤘",
	IconMore:     "⋯",
	IconFollow:   "Follow",
	IconSearch:   "🔍︎",
}

// Icon renders an icon by name. Emoji placeholder.
// TODO: SVG atlas lookup.
func Icon(name IconName) g.Node {
	return g.Text(iconGlyphs[name])
}
