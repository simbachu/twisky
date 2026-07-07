package ui

import (
	"net/url"
	"time"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func RepostMeta(repostedBy AuthorInfo) g.Node {
	return P(
		g.Attr("class", "post-meta post-meta-repost"),
		g.Text("Reposted by "),
		A(
			g.Attr("href", "/"+repostedBy.Handle),
			g.Attr("class", "post-meta-handle"),
			g.Attr("style", "pointer-events: auto"),
			g.Text("@"+repostedBy.Handle),
		),
	)
}

func ReplyMeta(parent AuthorInfo, parentPostID string) g.Node {
	href := "/" + parent.Handle + "/post/" + url.PathEscape(parentPostID)
	return P(
		g.Attr("class", "post-meta post-meta-reply"),
		A(
			g.Attr("href", href),
			g.Attr("class", "post-meta-handle"),
			g.Attr("style", "pointer-events: auto"),
			g.Text("⤷ Reply to @"+parent.Handle),
		),
	)
}

func PostHeader(author AuthorInfo, createdAt, now time.Time, repostedBy, replyParent *AuthorInfo, replyParentPostID string) g.Node {
	children := []g.Node{
		g.Attr("class", "post-header"),
	}
	if repostedBy != nil {
		children = append(children, RepostMeta(*repostedBy))
	}
	children = append(children, postBylineContent(author, createdAt, now))
	if replyParent != nil && replyParentPostID != "" {
		children = append(children, ReplyMeta(*replyParent, replyParentPostID))
	}
	return Header(children...)
}

func postBylineContent(author AuthorInfo, createdAt, now time.Time) g.Node {
	return Div(
		g.Attr("class", "byline"),
		Avatar(author),
		Span(
			g.Attr("class", "byline-meta"),
			AuthorLink(author),
			Timestamp(createdAt, now),
		),
	)
}
