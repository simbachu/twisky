package profile

import (
	"net/url"
	"time"

	feedcomponent "github.com/simbachu/twisky/internal/components/feed"
	"github.com/simbachu/twisky/internal/components/page"
	postcomponent "github.com/simbachu/twisky/internal/components/post"
	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
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
			profileStats(view),
			g.If(view.Description != "", profileDescription(view)),
			maybePinnedPost(view.PinnedPostMaybe, now),
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

func profileStats(view profilequery.ProfileView) g.Node {
	return P(
		ui.FuzzyNumber(view.Followers), g.Text(" followers · "),
		ui.FuzzyNumber(view.Following), g.Text(" following · "),
		ui.FuzzyNumber(view.Posts), g.Text(" posts"),
	)
}

func profileDescription(view profilequery.ProfileView) g.Node {
	if len(view.DescriptionSegments) == 0 {
		return P(g.Text(view.Description))
	}
	return P(ui.RichText(view.DescriptionSegments))
}

func maybePinnedPost(maybe *feedquery.PostView, now time.Time) g.Node {
	if maybe == nil {
		return nil
	}
	return pinnedPost(*maybe, now)
}

func pinnedPost(view feedquery.PostView, now time.Time) g.Node {
	href := "/" + view.AuthorHandle + "/post/" + url.PathEscape(view.ID)
	return Section(
		g.Attr("class", "profile-pinned"),
		P(g.Attr("class", "profile-pinned-label"), g.Text("Pinned")),
		Div(
			g.Attr("class", "feed-item"),
			A(
				g.Attr("href", href),
				g.Attr("aria-label", "View post"),
			),
			Div(postcomponent.InsetPost(&view, now)),
		),
	)
}
