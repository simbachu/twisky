package home

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FeedHeader renders the home feed title and compose summoner.
func FeedHeader() g.Node {
	return Header(
		g.Attr("class", "feed-header"),
		H1(g.Text("Home")),
		A(
			g.Attr("href", "/my/posts/new"),
			g.Attr("class", "compose-summoner"),
			g.Attr("data-compose-open", ""),
			g.Text("What's on your mind?"),
		),
	)
}
