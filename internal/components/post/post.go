package post

import (
	"net/url"
	"strconv"
	"time"

	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/richtext"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Post(view feedquery.PostView) g.Node {
	return Article(g.Attr("class", "post"), g.Attr("id", "post-"+url.PathEscape(view.ID)),
		Header(
			A(g.Attr("href", "/"+view.AuthorHandle), g.Attr("style", "pointer-events: auto"),
				g.If(view.AuthorAvatar != "", Figure(
					Img(
						g.Attr("src", view.AuthorAvatar),
						g.Attr("alt", view.AuthorDisplayName),
					)))),
			authorLink(view),
			g.If(!view.CreatedAt.IsZero(), Span(
				g.Text(" · "),
				Time(DateTime(view.CreatedAt.UTC().Format(time.RFC3339)),
					g.Text(formatPostTime(view.CreatedAt)),
				),
			)),
		),
		postText(view),
		g.If(len(view.Images) > 0, Figure(
			g.Group(g.Map(view.Images, postImage)),
		)),
		quotedInset(view.QuotedPostMaybe),
		Footer(
			Nav(
				Ul(
					Li(Button(g.Text("🗪"), g.Attr("aria-label", "Replies"), g.If(view.ReplyCount > 0, g.Text(strconv.Itoa(view.ReplyCount))))),   // Replies
					Li(Button(g.Text("🔁"), g.Attr("aria-label", "Reposts"), g.If(view.RepostCount > 0, g.Text(strconv.Itoa(view.RepostCount))))), // Reposts
					Li(Button(g.Text("👍"), g.Attr("aria-label", "Likes"), g.If(view.LikeCount > 0, g.Text(strconv.Itoa(view.LikeCount))))),       // Likes
					// Bookmark button
					// Share button
					// More options button
				),
			),
		),
	)
}

func quotedInset(maybe *feedquery.PostView) g.Node {
	if maybe == nil {
		return nil
	}
	return InsetPost(maybe)
}

// Condensed post view with Author, Icon, Text and images
func InsetPost(view *feedquery.PostView) g.Node {
	return Article(g.Attr("class", "post inset-post"), g.Attr("id", "inset-post-"+url.PathEscape(view.ID)),
		Header(
			g.If(view.AuthorAvatar != "", Figure(
				FigCaption(g.Text(view.AuthorDisplayName)),
				Img(
					g.Attr("src", view.AuthorAvatar),
					g.Attr("alt", view.AuthorDisplayName),
					g.Attr("height", "100"),
					g.Attr("width", "100"),
				))),
			authorLink(*view),
			g.If(!view.CreatedAt.IsZero(), Time(
				DateTime(view.CreatedAt.UTC().Format(time.RFC3339)),
				g.Text(formatPostTime(view.CreatedAt)),
			)),
		),
		postText(*view),
		g.If(len(view.Images) > 0, Figure(
			g.Group(g.Map(view.Images, postImage)),
		)),
	)
}

// authorLink renders the post byline: display name and @handle as separate spans
// inside one profile link. When the display name equals the handle (no custom
// name set), only the @handle span is shown to avoid "handle @handle" duplication.
func authorLink(view feedquery.PostView) g.Node {
	children := []g.Node{g.Attr("href", "/"+view.AuthorHandle), g.Attr("class", "post-author"), g.Attr("style", "pointer-events: auto")}
	if view.AuthorDisplayName != view.AuthorHandle {
		children = append(children, Span(g.Attr("class", "author-name"), g.Text(view.AuthorDisplayName)))
		children = append(children, Span(g.Attr("class", "author-handle"), g.Text(" @"+view.AuthorHandle)))
	} else {
		children = append(children, Span(g.Attr("class", "author-handle"), g.Text("@"+view.AuthorHandle)))
	}
	return A(children...)
}

func formatPostTime(createdAt time.Time) string {
	return createdAt.UTC().Format("Jan 2, 2006, 3:04 PM UTC")
}

func postText(view feedquery.PostView) g.Node {
	if len(view.TextSegments) == 0 {
		return P(g.Text(view.Text))
	}
	return P(g.Group(g.Map(view.TextSegments, segmentNode)))
}

func segmentNode(segment richtext.Segment) g.Node {
	switch segment.Kind {
	case richtext.Tag:
		return A(
			g.Attr("href", "/tagged/"+url.PathEscape(segment.Tag)),
			g.Attr("style", "pointer-events: auto"),
			Span(g.Attr("class", "facet-tag"), g.Text(segment.Text)),
		)
	case richtext.Mention:
		return A(
			g.Attr("href", "/"+url.PathEscape(segment.Mention)),
			g.Attr("style", "pointer-events: auto"),
			Span(g.Attr("class", "facet-mention"), g.Text(segment.Text)),
		)
	case richtext.Link:
		return A(
			g.Attr("href", segment.URI),
			g.Attr("target", "_blank"),
			g.Attr("rel", "noopener noreferrer"),
			g.Attr("style", "pointer-events: auto"),
			Span(g.Attr("class", "facet-link"), g.Text(segment.Text)),
		)
	default:
		return g.Text(segment.Text)
	}
}

func postImage(image feedquery.ImageView) g.Node {
	attrs := []g.Node{
		g.Attr("src", image.Thumb),
		g.Attr("alt", image.Alt),
		g.Attr("loading", "lazy"),
	}
	if image.Fullsize != "" {
		attrs = append(attrs, g.Attr("srcset", image.Thumb+" 1000w, "+image.Fullsize+" 2000w"))
	}
	if image.Width > 0 && image.Height > 0 {
		attrs = append(attrs,
			g.Attr("width", strconv.Itoa(image.Width)),
			g.Attr("height", strconv.Itoa(image.Height)),
		)
	}
	return Img(attrs...)
}
