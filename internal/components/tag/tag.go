package tag

import (
	"net/url"
	"time"

	feedcomponent "github.com/simbachu/twisky/internal/components/feed"
	"github.com/simbachu/twisky/internal/components/page"
	tagquery "github.com/simbachu/twisky/internal/query/tag"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Tag(view tagquery.TagView, now time.Time) g.Node {
	title := "#" + view.Tag
	feedURL := "/tagged/" + url.PathEscape(view.Tag)

	children := []g.Node{
		Header(
			H1(g.Text(title)),
		),
	}
	if len(view.Feed.Posts) > 0 {
		children = append(children, feedcomponent.NewPostsPoll(feedURL, view.Feed.Posts[0].ID))
	}
	children = append(children, feedcomponent.Feed(view.Feed, now, feedURL))

	return page.Page(
		"Viewing tag: "+title,
		"Viewing posts tagged with "+title,
		children...,
	)
}
