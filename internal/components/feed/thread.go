package feed

import (
	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FeedThread renders a reply with its parent in short-thread form for feeds.
func FeedThread(parent, child feedquery.PostView) g.Node {
	return Div(
		g.Attr("class", "feed-thread"),
		post.InsetPost(&parent),
		post.Post(child),
	)
}
