package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MiniProfile renders a compact profile link (avatar, name, handle).
func MiniProfile(author AuthorInfo) g.Node {
	children := []g.Node{
		g.Attr("href", "/"+author.Handle),
		g.Attr("class", "mini-profile"),
	}
	if glyph := AvatarGlyph(author.Avatar, author.Handle, author.DID, author.IsLabeler); glyph != nil {
		children = append(children, Span(
			g.Attr("class", "byline-avatar"),
			glyph,
		))
	}
	children = append(children, Span(
		g.Attr("class", "mini-profile-info"),
		AuthorName(author),
		AuthorHandle(author),
	))
	return A(children...)
}
