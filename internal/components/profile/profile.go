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
	feedURL := "/" + view.Handle
	if view.Tab == profilequery.TabMedia {
		feedURL += "/media"
	}

	children := []g.Node{
		Header(
			g.Attr("class", "profile-header"),
			ui.Avatar(author),
			H1(ui.AuthorName(author)),
			H2(ui.AuthorHandle(author)),
			g.If(view.Description != "", profileDescription(view)),
			P(g.Textf("%d followers · %d following · %d posts", view.Followers, view.Following, view.Posts)),
		),
		ui.TabNav("Profile", []ui.TabItem{
			{Label: "Posts", Href: "/" + view.Handle, Current: view.Tab == profilequery.TabPosts},
			{Label: "Media", Href: "/" + view.Handle + "/media", Current: view.Tab == profilequery.TabMedia},
		}),
	}
	if len(view.Feed.Posts) > 0 {
		children = append(children, feedcomponent.NewPostsPoll(feedURL, view.Feed.Posts[0].ID))
	}
	children = append(children, feedcomponent.Feed(view.Feed, now, feedURL))

	return page.Page(
		"Viewing profile: "+view.DisplayName,
		"Viewing the profile of "+view.DisplayName,
		children...,
	)
}

func profileDescription(view profilequery.ProfileView) g.Node {
	if len(view.DescriptionSegments) == 0 {
		return P(g.Text(view.Description))
	}
	return P(ui.RichText(view.DescriptionSegments))
}
