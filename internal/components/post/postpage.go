package post

import (
	"net/url"
	"time"

	"github.com/simbachu/twisky/internal/components/page"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func PostPage(view feedquery.PostPageView, now time.Time) g.Node {
	return page.Page(
		"Post by "+view.Post.AuthorDisplayName,
		"Viewing a post by "+view.Post.AuthorDisplayName,
		g.If(len(view.Ancestors) > 0, Section(
			g.Attr("class", "thread-ancestors"),
			Ul(g.Group(g.Map(view.Ancestors, func(postView feedquery.PostView) g.Node {
				return ancestorItem(postView, now)
			}))),
		)),
		Section(g.Attr("class", "thread-root"), threadRootPost(view.Post, now)),
		g.If(len(view.Replies) > 0, Section(
			g.Attr("class", "thread-replies"),
			Ul(g.Group(g.Map(view.Replies, func(node feedquery.ThreadNodeView) g.Node {
				return replyItem(node, now)
			}))),
		)),
	)
}

func threadRootPost(view feedquery.PostView, now time.Time) g.Node {
	if view.Moderation.Filtered {
		return P(g.Text("Post hidden by moderation"))
	}
	return Post(view, now)
}

func ancestorItem(postView feedquery.PostView, now time.Time) g.Node {
	if postView.Moderation.Filtered {
		return Li(P(g.Text("Post hidden by moderation")))
	}
	return Li(Post(postView, now))
}

func replyItem(node feedquery.ThreadNodeView, now time.Time) g.Node {
	if node.Unavailable {
		return Li(P(g.Text("Post unavailable")))
	}
	if node.Post.Moderation.Filtered {
		return Li(P(g.Text("Post hidden by moderation")))
	}

	href := "/" + node.Post.AuthorHandle + "/post/" + url.PathEscape(node.Post.ID)
	return Li(
		A(
			g.Attr("href", href),
			g.Attr("style", "pointer-events: auto"),
			Post(node.Post, now),
		),
		g.If(len(node.Replies) > 0, Ul(g.Group(g.Map(node.Replies, func(child feedquery.ThreadNodeView) g.Node {
			return replyItem(child, now)
		})))),
	)
}
