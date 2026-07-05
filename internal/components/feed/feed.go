package feed

import (
	"net/url"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Feed renders a list of posts. Use to compose a page.
// Each list item is covered by an anchor overlay so users can open the post by
// clicking anywhere that is not another link (author, tag, mention, etc.).
func Feed(view feedquery.FeedView) g.Node {
	return Ul(g.Group(g.Map(view.Posts, func(postView feedquery.PostView) g.Node {
		href := "/" + postView.AuthorHandle + "/post/" + url.PathEscape(postView.ID)
		return Li(g.Attr("style", "position: relative"),
			Div(g.Attr("style", "pointer-events: none"), post.Post(postView)),
			A(
				g.Attr("href", href),
				g.Attr("style", "position: absolute; inset: 0"),
				g.Attr("aria-label", "View post"),
			),
		)
	})))
}
