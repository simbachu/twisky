package tag

import (
	"time"

	feedcomponent "github.com/simbachu/twisky/internal/components/feed"
	"github.com/simbachu/twisky/internal/components/page"
	tagquery "github.com/simbachu/twisky/internal/query/tag"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Tag(view tagquery.TagView, now time.Time) g.Node {
	title := "#" + view.Tag

	return page.Page(
		"Viewing tag: "+title,
		"Viewing posts tagged with "+title,
		Header(
			H1(g.Text(title)),
		),
		feedcomponent.Feed(view.Feed, now),
	)
}
