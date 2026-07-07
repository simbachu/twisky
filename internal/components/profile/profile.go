package profile

import (
	"time"

	feedcomponent "github.com/simbachu/twisky/internal/components/feed"
	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	profilequery "github.com/simbachu/twisky/internal/query/profile"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Profile(view profilequery.ProfileView, now time.Time) g.Node {
	author := ui.AuthorInfo{
		Handle:      view.Handle,
		DisplayName: view.DisplayName,
		Avatar:      view.Avatar,
	}

	return page.Page(
		"Viewing profile: "+view.DisplayName,
		"Viewing the profile of "+view.DisplayName,
		Header(
			g.Attr("class", "profile-header"),
			ui.Avatar(author),
			H1(ui.AuthorName(author)),
			H2(ui.AuthorHandle(author)),
			g.If(view.Description != "", P(g.Text(view.Description))),
			P(g.Textf("%d followers · %d following · %d posts", view.Followers, view.Following, view.Posts)),
		),
		ui.TabNav("Profile", []ui.TabItem{
			{Label: "Posts", Href: "/" + view.Handle, Current: view.Tab == profilequery.TabPosts},
			{Label: "Media", Href: "/" + view.Handle + "/media", Current: view.Tab == profilequery.TabMedia},
		}),
		feedcomponent.Feed(view.Feed, now),
	)
}
