package ui

import (
	"time"

	"github.com/simbachu/twisky/internal/actor"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type AuthorInfo struct {
	Handle      string
	DisplayName string
	DID         string
	Avatar      string
	IsLabeler   bool
	BlurAvatar  bool
	AvatarText  string
}

func avatarGlyph(avatarURL, handle, did string, isLabeler bool, alt string) g.Node {
	if avatarURL != "" {
		return Img(
			g.Attr("src", avatarURL),
			g.Attr("alt", alt),
		)
	}
	if emoji := actor.AvatarFallbackEmoji(handle, did, isLabeler); emoji != "" {
		return Span(g.Attr("class", "byline-avatar-emoji"), g.Text(emoji))
	}
	return nil
}

func Avatar(author AuthorInfo) g.Node {
	if author.BlurAvatar {
		message := author.AvatarText
		if message == "" {
			message = "Content warning"
		}
		attrs := []g.Node{
			g.Attr("class", "byline-avatar byline-avatar-moderated"),
			g.Attr("title", message),
		}
		if author.IsLabeler {
			attrs = append(attrs, g.Attr("data-kind", "labeler"))
		}
		attrs = append(attrs, g.Text(""))
		return Span(attrs...)
	}
	glyph := avatarGlyph(author.Avatar, author.Handle, author.DID, author.IsLabeler, author.DisplayName)
	if glyph == nil {
		return nil
	}
	attrs := []g.Node{
		g.Attr("href", "/"+author.Handle),
		g.Attr("class", "byline-avatar"),
		g.Attr("style", "pointer-events: auto"),
	}
	if author.IsLabeler {
		attrs = append(attrs, g.Attr("data-kind", "labeler"))
	}
	attrs = append(attrs, glyph)
	return A(attrs...)
}

// AvatarGlyph renders a small avatar image or fallback emoji for inline use.
func AvatarGlyph(avatarURL, handle, did string, isLabeler bool) g.Node {
	return avatarGlyph(avatarURL, handle, did, isLabeler, "")
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
	return postBylineContent(author, createdAt, now, true, false)
}
