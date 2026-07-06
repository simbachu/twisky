package post

import (
	"net/url"

	"github.com/simbachu/twisky/internal/components/page"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func PostPage(view feedquery.PostPageView) g.Node {
	return page.Page(
		"Post by "+view.Post.AuthorDisplayName,
		"Viewing a post by "+view.Post.AuthorDisplayName,
		g.If(len(view.Ancestors) > 0, Section(
			g.Attr("class", "thread-ancestors"),
			Ul(g.Group(g.Map(view.Ancestors, ancestorItem))),
		)),
		Section(g.Attr("class", "thread-root"), Post(view.Post)),
		g.If(len(view.Replies) > 0, Section(
			g.Attr("class", "thread-replies"),
			Ul(g.Group(g.Map(view.Replies, replyItem))),
		)),
	)
}

func ancestorItem(postView feedquery.PostView) g.Node {
	return Li(Post(postView))
}

func replyItem(node feedquery.ThreadNodeView) g.Node {
	if node.Unavailable {
		return Li(P(g.Text("Post unavailable")))
	}

	href := "/" + node.Post.AuthorHandle + "/post/" + url.PathEscape(node.Post.ID)
	return Li(
		A(
			g.Attr("href", href),
			g.Attr("style", "pointer-events: auto"),
			Post(node.Post),
		),
		g.If(len(node.Replies) > 0, Ul(g.Group(g.Map(node.Replies, replyItem)))),
	)
}
