package feed

import (
	"fmt"
	"net/url"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Feed renders a feed list with htmx infinite-scroll sentinel. Use to compose a page.
func Feed(view feedquery.FeedView, now time.Time, feedURL string) g.Node {
	return Ul(
		g.Attr("id", "feed-list"),
		FeedItems(view, now, feedURL),
	)
}

// FeedItems renders post list items plus a trailing sentinel for infinite scroll.
// Used for both full pages and cursor-fragment responses.
func FeedItems(view feedquery.FeedView, now time.Time, feedURL string) g.Node {
	nodes := g.Map(view.Posts, func(postView feedquery.PostView) g.Node {
		return feedItem(postView, now)
	})
	if view.NextCursor != "" {
		nodes = append(nodes, feedSentinel(feedURL, view.NextCursor))
	}
	return nodes
}

// PrependItems renders bare list items for prepending new posts into the feed.
func PrependItems(posts []feedquery.PostView, now time.Time) g.Group {
	return g.Map(posts, func(postView feedquery.PostView) g.Node {
		return feedItem(postView, now)
	})
}

// NewPostsPoll renders the periodic new-posts poller above the feed.
func NewPostsPoll(feedURL, sinceID string) g.Node {
	return Div(
		g.Attr("id", "new-posts-slot"),
		g.Attr("class", "new-posts-slot"),
		g.Attr("hx-get", feedURL+"?since="+url.QueryEscape(sinceID)),
		g.Attr("hx-trigger", "every 20s"),
		g.Attr("hx-swap", "innerHTML"),
	)
}

// NewPostsPollOOB resets the poller after prepending new posts.
func NewPostsPollOOB(feedURL, sinceID string) g.Node {
	return Div(
		g.Attr("id", "new-posts-slot"),
		g.Attr("class", "new-posts-slot"),
		g.Attr("hx-get", feedURL+"?since="+url.QueryEscape(sinceID)),
		g.Attr("hx-trigger", "every 20s"),
		g.Attr("hx-swap", "innerHTML"),
		g.Attr("hx-swap-oob", "true"),
	)
}

// NewPostsBanner renders a button to load new posts, or nil when count is zero.
func NewPostsBanner(count int, feedURL, sinceID string) g.Node {
	if count == 0 {
		return nil
	}
	label := "Show 1 new post"
	if count != 1 {
		label = fmt.Sprintf("Show %d new posts", count)
	}
	return Button(
		g.Attr("type", "button"),
		g.Attr("class", "new-posts-button"),
		g.Attr("hx-get", feedURL+"?refresh="+url.QueryEscape(sinceID)),
		g.Attr("hx-target", "#feed-list"),
		g.Attr("hx-swap", "afterbegin"),
		g.Text(label),
	)
}

func feedItem(postView feedquery.PostView, now time.Time) g.Node {
	href := "/" + postView.AuthorHandle + "/post/" + url.PathEscape(postView.ID)
	return Li(g.Attr("class", "feed-item"),
		A(
			g.Attr("href", href),
			g.Attr("class", "feed-item-overlay"),
			g.Attr("aria-label", "View post"),
		),
		Div(g.Attr("class", "feed-item-content"), post.Post(postView, now)),
	)
}

func feedSentinel(feedURL, cursor string) g.Node {
	return Li(
		g.Attr("class", "feed-sentinel"),
		g.Attr("hx-get", feedURL+"?cursor="+url.QueryEscape(cursor)),
		g.Attr("hx-trigger", "revealed"),
		g.Attr("hx-swap", "outerHTML"),
	)
}
