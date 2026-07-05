package profile

import (
	feedcomponent "github.com/simbachu/twisky/internal/components/feed"
	"github.com/simbachu/twisky/internal/components/page"
	profilequery "github.com/simbachu/twisky/internal/query/profile"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Profile(view profilequery.ProfileView) g.Node {
	title := view.DisplayName
	if title == "" {
		title = view.Handle
	}

	return page.Page(
		"Viewing profile: "+title,
		"Viewing the profile of "+title,
		Header(
			H1(g.Text(title)),
			Img(
				g.Attr("src", view.Avatar),
				g.Attr("alt", view.DisplayName),
				g.Attr("height", "100"),
				g.Attr("width", "100"),
			),
			H2(g.Text("@"+view.Handle)),
			g.If(view.Description != "", P(g.Text(view.Description))),
			P(g.Textf("%d followers · %d following · %d posts", view.Followers, view.Following, view.Posts)),
		),
		Nav(
			Ul(
				Li(A(
					g.Attr("href", "/"+view.Handle),
					g.Text("Posts"),
					g.If(view.Tab == profilequery.TabPosts, g.Attr("aria-current", "page")),
				)),
				Li(A(
					g.Attr("href", "/"+view.Handle+"/media"),
					g.Text("Media"),
					g.If(view.Tab == profilequery.TabMedia, g.Attr("aria-current", "page")),
				)),
			),
		),
		feedcomponent.Feed(view.Feed),
	)
}
