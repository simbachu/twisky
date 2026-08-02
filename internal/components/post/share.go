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
	attrs := []g.Node{
		g.Attr("type", "button"),
		g.Attr("data-copy-url", copyURL),
		g.Attr("aria-label", label),
	}
	switch label {
	case "Copy Twisky link":
		attrs = append(attrs, g.Text("🔗"))
	case "Copy Bluesky link":
		attrs = append(attrs,
			g.Attr("class", ui.ActionClass(ui.IconBluesky)),
			g.Attr("data-copy-feedback", "icon"),
			ui.Icon(ui.IconBluesky),
		)
	default:
		attrs = append(attrs, g.Text(label))
	}
	return Button(g.Group(attrs))
}

func shareGroup(view feedquery.PostView) g.Node {
	twiskyPath := postHref(view)
	bskyURL := blueskyPostPageURL(view.AuthorHandle, view.ID)
	return ui.SegmentedShell("Share", "post-share-group",
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
