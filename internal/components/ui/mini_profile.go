package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MiniProfile renders compact author chrome (avatar, name, handle) for the given mode.
// It is layout only — it does not wrap itself in a profile link.
func MiniProfile(author AuthorInfo, mode AuthorInteractive) g.Node {
	children := []g.Node{
		g.Attr("class", "mini-profile"),
	}
	if avatar := ActionableAvatar(author, mode); avatar != nil {
		children = append(children, avatar)
	}
	children = append(children, Span(
		g.Attr("class", "mini-profile-info"),
		AuthorName(author, mode),
		AuthorHandle(author, mode),
	))
	return Span(children...)
}
