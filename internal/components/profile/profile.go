package profile

import (
	"time"

	feedcomponent "github.com/simbachu/twisky/internal/components/feed"
	"github.com/simbachu/twisky/internal/components/page"
	postcomponent "github.com/simbachu/twisky/internal/components/post"
	"github.com/simbachu/twisky/internal/components/ui"
	"github.com/simbachu/twisky/internal/moderation"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	profilequery "github.com/simbachu/twisky/internal/query/profile"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Profile(view profilequery.ProfileView, now time.Time, suggested []ui.AuthorInfo, auth page.AuthChrome, publicBaseURL string) g.Node {
	author := ui.AuthorInfo{
		Handle:      view.Handle,
		DisplayName: view.DisplayName,
		DID:         view.DID,
		Avatar:      view.Avatar,
		IsLabeler:   view.IsLabeler,
		BlurAvatar:  view.AvatarModeration.BlurAvatar,
		AvatarText:  view.AvatarModeration.AvatarText,
	}
	feedURL := "/" + view.Handle
	if view.Tab == profilequery.TabMedia {
		feedURL += "/media"
	}

	children := []g.Node{
		Header(
			g.Attr("class", "profile"),
			profileBanner(view.Banner),
			ui.Avatar(author),
			H1(ui.AuthorName(author)),
			H2(ui.AuthorHandle(author)),
			profileStats(view),
			profileLabels(view.Labels),
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
		profilePageMeta(view, publicBaseURL),
		suggested,
		auth,
		children...,
	)
}

func profileBanner(url string) g.Node {
	if url == "" {
		return Figure()
	}
	return Figure(
		Img(
			g.Attr("src", url),
			g.Attr("alt", ""),
		),
	)
}

func profileStats(view profilequery.ProfileView) g.Node {
	return P(
		ui.FuzzyNumber(view.Followers), g.Text(" followers · "),
		ui.FuzzyNumber(view.Following), g.Text(" following · "),
		ui.FuzzyNumber(view.Posts), g.Text(" posts"),
	)
}

func profileLabels(labels []moderation.ProfileLabelView) g.Node {
	if len(labels) == 0 {
		return nil
	}

	pills := make([]g.Node, len(labels))
	for i, label := range labels {
		children := []g.Node{
			g.Attr("class", "iface-pill"),
			g.Attr("href", "/"+label.LabelerHandle),
			g.Attr("title", label.Labeler+": "+label.Message),
		}
		if glyph := ui.AvatarGlyph(label.LabelerAvatar, label.LabelerHandle, label.LabelerDID, label.LabelerIsLabeler); glyph != nil {
			children = append(children, glyph)
		}
		children = append(children, Span(g.Text(label.Message)))
		pills[i] = A(g.Group(children))
	}

	return Div(
		g.Attr("class", "profile-labels"),
		g.Group(pills),
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
	return Section(
		g.Attr("class", "profile-pinned"),
		P(g.Attr("class", "profile-pinned-label"), g.Text("Pinned")),
		postcomponent.ClickableInset(&view, now, "View post"),
	)
}
