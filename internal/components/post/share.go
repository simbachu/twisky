package post

import (
	"net/url"

	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func blueskyPostPageURL(handle, postID string) string {
	return "https://bsky.app/profile/" + handle + "/post/" + url.PathEscape(postID)
}

func shareCopyButton(label, copyURL string) g.Node {
	return Button(
		g.Attr("type", "button"),
		g.Attr("data-copy-url", copyURL),
		g.Attr("aria-label", label),
		g.Text(shareCopyLabel(label)),
	)
}

func shareCopyLabel(label string) string {
	switch label {
	case "Copy Twisky link":
		return "🔗"
	case "Copy Bluesky link":
		return "🦋"
	default:
		return label
	}
}

func shareGroup(view feedquery.PostView) g.Node {
	twiskyPath := postHref(view)
	bskyURL := blueskyPostPageURL(view.AuthorHandle, view.ID)
	return Menu(
		g.Attr("class", "iface-segmented post-share-group"),
		g.Attr("aria-label", "Share"),
		g.Attr("role", "group"),
		Li(
			g.Attr("class", "post-share-trigger"),
			Button(
				g.Attr("type", "button"),
				g.Attr("class", "post-share-open"),
				g.Attr("aria-label", "Share"),
				g.Attr("aria-expanded", "false"),
				g.Attr("aria-haspopup", "menu"),
				ui.Icon(ui.IconShare),
			),
		),
		Li(
			g.Attr("class", "post-share-option"),
			shareCopyButton("Copy Twisky link", twiskyPath),
		),
		Li(
			g.Attr("class", "post-share-option"),
			shareCopyButton("Copy Bluesky link", bskyURL),
		),
	)
}
