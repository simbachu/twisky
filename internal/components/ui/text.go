package ui

import (
	"net/url"

	"github.com/simbachu/twisky/internal/richtext"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func RichText(segments []richtext.Segment) g.Node {
	return g.Group(g.Map(segments, SegmentNode))
}

func SegmentNode(segment richtext.Segment) g.Node {
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
			Span(
				g.Attr("class", "facet-link"),
				g.Text(segment.Text),
				g.Text("🡕"),
			),
		)
	default:
		return g.Text(segment.Text)
	}
}
