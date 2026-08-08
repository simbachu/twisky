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

func Home(view homequery.HomeView, now time.Time, suggested []ui.AuthorInfo, accounts ui.AccountMenuView, publicBaseURL string) g.Node {
	feedURL := view.Path
	if feedURL == "" {
		feedURL = "/"
	}
	tabs := make([]ui.TabItem, len(view.Tabs))
	for index, tab := range view.Tabs {
		tabs[index] = ui.TabItem{
			Label: tab.Label, Href: tab.Href, Current: tab.Current,
		}
	}
	children := []g.Node{
		Header(
			H1(g.Text("Home")),
		),
		ui.TabNav("Home feeds", tabs),
	}
	if len(view.Feed.Posts) > 0 {
		children = append(children, feedcomponent.NewPostsPoll(feedURL, view.Feed.Posts[0].ID))
	}
	children = append(children, feedcomponent.Feed(view.Feed, now, feedURL))

	return page.Page(
		homePageMeta(view, publicBaseURL),
		suggested,
		accounts,
		children...,
	)
}

func homePageMeta(view homequery.HomeView, publicBaseURL string) page.PageMeta {
	title := view.Title
	if title == "" || title == "Following" {
		title = "Home"
	}
	path := view.Path
	if path == "" {
		path = "/"
	}
	return page.PageMeta{
		Title:        title + " · Twisky",
		Description:  title + " feed on Twisky.",
		CanonicalURL: page.AbsoluteURL(publicBaseURL, path),
		Path:         "/",
		OGType:       "website",
	}
}
