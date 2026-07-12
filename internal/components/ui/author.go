package ui

import (
	"time"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type AuthorInfo struct {
	Handle      string
	DisplayName string
	Avatar      string
}

func Avatar(author AuthorInfo) g.Node {
	if author.Avatar == "" {
		return nil
	}
	return A(
		g.Attr("href", "/"+author.Handle),
		g.Attr("class", "byline-avatar"),
		g.Attr("style", "pointer-events: auto"),
		Img(
			g.Attr("src", author.Avatar),
			g.Attr("alt", author.DisplayName),
		),
	)
}

func AuthorName(author AuthorInfo) g.Node {
	return Span(g.Attr("class", "author-name"), g.Text(author.DisplayName))
}

func AuthorHandle(author AuthorInfo) g.Node {
	return Span(g.Attr("class", "author-handle"), g.Text("@"+author.Handle))
}

func AuthorLink(author AuthorInfo) g.Node {
	children := []g.Node{
		g.Attr("href", "/"+author.Handle),
		g.Attr("style", "pointer-events: auto"),
	}
	if author.DisplayName != author.Handle {
		children = append(children,
			AuthorName(author),
			Span(g.Attr("class", "author-handle"), g.Text(" @"+author.Handle)),
		)
	} else {
		children = append(children, AuthorHandle(author))
	}
	return A(children...)
}

func PostByline(author AuthorInfo, createdAt, now time.Time) g.Node {
	return postBylineContent(author, createdAt, now)
}
