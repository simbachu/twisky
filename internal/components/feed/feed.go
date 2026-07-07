package feed

import (
	"net/url"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Feed renders a list of posts. Use to compose a page.
// Each list item is covered by an anchor overlay so users can open the post by
// clicking anywhere that is not another link (author, tag, mention, etc.).
func Feed(view feedquery.FeedView, now time.Time) g.Node {
	return Ul(g.Group(g.Map(view.Posts, func(postView feedquery.PostView) g.Node {
		href := "/" + postView.AuthorHandle + "/post/" + url.PathEscape(postView.ID)
		return Li(g.Attr("class", "feed-item"),
			A(
				g.Attr("href", href),
				g.Attr("class", "feed-item-overlay"),
				g.Attr("aria-label", "View post"),
			),
			Div(g.Attr("class", "feed-item-content"), post.Post(postView, now)),
		)
	})))
}
