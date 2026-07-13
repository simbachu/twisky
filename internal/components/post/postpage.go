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
		g.Group{
			g.If(view.HasAncestors, postPageAncestorsSlot(view.Post)),
			postPageRoot(view.Post, view.Replies, now),
		},
	)
}

func PostPageAncestors(view feedquery.PostPageView, now time.Time) g.Node {
	return postPageAncestorsContent(view.Ancestors, now)
}

func postPageAncestorsSlot(post feedquery.PostView) g.Node {
	href := "/" + post.AuthorHandle + "/post/" + url.PathEscape(post.ID)
	return Section(
		g.Attr("id", "post-page-ancestors"),
		g.Attr("class", "post-page-ancestors"),
		g.Attr("aria-label", "Thread context"),
		g.Attr("hx-get", href+"?ancestors=1"),
		g.Attr("hx-trigger", "twiskyAncestors"),
		g.Attr("hx-swap", "innerHTML"),
	)
}

func postPageAncestorsContent(ancestors []feedquery.AncestorNodeView, now time.Time) g.Node {
	return g.Group(g.Map(furthestFirstAncestors(ancestors), func(ancestor feedquery.AncestorNodeView) g.Node {
		return ancestorItem(ancestor, now)
	}))
}

func ancestorItem(node feedquery.AncestorNodeView, now time.Time) g.Node {
	if node.Unavailable {
		return P(g.Text("Post unavailable"))
	}
	if node.Post.Moderation.Filtered {
		return P(g.Text("Post hidden by moderation"))
	}
	return Post(node.Post, now)
}

func furthestFirstAncestors(ancestors []feedquery.AncestorNodeView) []feedquery.AncestorNodeView {
	reversed := make([]feedquery.AncestorNodeView, len(ancestors))
	for i, node := range ancestors {
		reversed[len(ancestors)-1-i] = node
	}
	return reversed
}

func postPageRoot(view feedquery.PostView, replies []feedquery.ThreadNodeView, now time.Time) g.Node {
	if view.Moderation.Filtered {
		return P(g.Text("Post hidden by moderation"))
	}
	var extra []g.Node
	if len(replies) > 0 {
		extra = append(extra, repliesList(replies, now))
	}
	return PostArticle(view, now, "post post-page", extra...)
}

func repliesList(replies []feedquery.ThreadNodeView, now time.Time) g.Node {
	return Ul(
		g.Attr("class", "post-replies"),
		g.Group(g.Map(replies, func(node feedquery.ThreadNodeView) g.Node {
			return replyItem(node, now)
		})),
	)
}

func replyItem(node feedquery.ThreadNodeView, now time.Time) g.Node {
	if node.Unavailable {
		return Li(P(g.Text("Post unavailable")))
	}
	if node.Post.Moderation.Filtered {
		return Li(P(g.Text("Post hidden by moderation")))
	}

	return Li(
		Post(node.Post, now),
		g.If(len(node.Replies) > 0, repliesList(node.Replies, now)),
	)
}
