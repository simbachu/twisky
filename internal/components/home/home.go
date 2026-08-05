package home

import (
	"time"

	feedcomponent "github.com/simbachu/twisky/internal/components/feed"
	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	homequery "github.com/simbachu/twisky/internal/query/home"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const feedURL = "/"

func Home(view homequery.HomeView, now time.Time, suggested []ui.AuthorInfo, accounts ui.AccountMenuView, publicBaseURL string) g.Node {
	children := []g.Node{
		Header(
			H1(g.Text("Home")),
		),
	}
	if len(view.Feed.Posts) > 0 {
		children = append(children, feedcomponent.NewPostsPoll(feedURL, view.Feed.Posts[0].ID))
	}
	children = append(children, feedcomponent.Feed(view.Feed, now, feedURL))

	return page.Page(
		homePageMeta(publicBaseURL),
		suggested,
		accounts,
		children...,
	)
}

func homePageMeta(publicBaseURL string) page.PageMeta {
	return page.PageMeta{
		Title:        "Home · Twisky",
		Description:  "Your following feed on Twisky.",
		CanonicalURL: page.AbsoluteURL(publicBaseURL, "/"),
		Path:         "/",
		OGType:       "website",
	}
}
