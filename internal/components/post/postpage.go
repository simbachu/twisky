package post

import (
	"net/url"
	"time"

	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func PostPage(view feedquery.PostPageView, now time.Time, suggested []ui.AuthorInfo, publicBaseURL string) g.Node {
	return page.Page(
		postPageMeta(view, publicBaseURL),
		suggested,
		g.Group{
			g.If(view.HasAncestors, postPageAncestorsSlot(view.Post)),
			postPageRoot(view.Post, view.Replies, now, view.ExplicitLive),
		},
	)
}

func PostPageAncestors(view feedquery.PostPageView, now time.Time) g.Node {
	return postPageAncestorsContent(view.Ancestors, now)
}

// RepliesRefreshFragment renders an out-of-band replacement of the top-level
// replies list when the thread contains posts the client does not yet know.
// Returns an empty group when every available reply ID is already in known.
func RepliesRefreshFragment(view feedquery.PostPageView, known map[string]bool, now time.Time) g.Node {
	if !feedquery.ThreadHasUnknown(view.Replies, known) {
		return g.Group{}
	}
	return repliesList(view.Replies, now, repliesRootID(view.Post.ID), true)
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

func postPageRoot(view feedquery.PostView, replies []feedquery.ThreadNodeView, now time.Time, explicitLive bool) g.Node {
	if view.Moderation.Filtered {
		return P(g.Text("Post hidden by moderation"))
	}
	live := explicitLive || autoStartLive(now.Sub(view.CreatedAt))
	// Always render a stable replies container so live refresh can OOB-swap
	// into it even when the page first loaded with zero replies.
	extra := []g.Node{repliesList(replies, now, repliesRootID(view.ID), false)}
	return PostArticle(view, now, "post post-page", true, live, extra...)
}

func repliesRootID(postID string) string {
	return "post-replies-" + url.PathEscape(postID)
}

// repliesList renders a reply tree. rootID is set only on the top-level list
// (swap target for live refresh); nested lists pass "". When oob is true the
// list carries hx-swap-oob for out-of-band replacement.
func repliesList(replies []feedquery.ThreadNodeView, now time.Time, rootID string, oob bool) g.Node {
	attrs := []g.Node{g.Attr("class", "post-replies")}
	if rootID != "" {
		attrs = append(attrs, g.Attr("id", rootID))
	}
	if oob {
		attrs = append(attrs, g.Attr("hx-swap-oob", "true"))
	}
	return Ul(
		g.Group(attrs),
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
		g.If(len(node.Replies) > 0, repliesList(node.Replies, now, "", false)),
	)
}
