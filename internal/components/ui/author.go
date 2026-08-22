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

// AuthorInteractive is how an author reference behaves.
type AuthorInteractive int

const (
	// AuthorStatic — paint only (profile header, account menu, supporting chrome).
	AuthorStatic AuthorInteractive = iota
	// AuthorLinked — navigates to profile; no hover peek (controls inside a peek card).
	AuthorLinked
	// AuthorPeekable — navigates + hover peek (timeline, recommendation discrete links).
	AuthorPeekable
)

func authorInteractiveAttr(mode AuthorInteractive) g.Node {
	switch mode {
	case AuthorLinked:
		return g.Attr("data-author-interactive", "linked")
	case AuthorPeekable:
		return g.Attr("data-author-interactive", "peekable")
	default:
		return nil
	}
}

func avatarGlyph(avatarURL, handle, did string, isLabeler bool, alt string) g.Node {
	if avatarURL != "" {
		return Img(
			g.Attr("src", avatarURL),
			g.Attr("alt", alt),
		)
	}
	if emoji := actor.AvatarFallbackEmoji(handle, did, isLabeler); emoji != "" {
		return Span(g.Attr("class", "profile-icon-emoji"), g.Text(emoji))
	}
	return nil
}

func profileIconClass(moderated bool) string {
	if moderated {
		return "profile-icon profile-icon-moderated"
	}
	return "profile-icon"
}

func moderatedAvatar(author AuthorInfo, mode AuthorInteractive) g.Node {
	message := author.AvatarText
	if message == "" {
		message = "Content warning"
	}
	attrs := []g.Node{
		g.Attr("class", profileIconClass(true)),
		g.Attr("title", message),
	}
	if author.IsLabeler {
		attrs = append(attrs, g.Attr("data-kind", "labeler"))
	}
	if attr := authorInteractiveAttr(mode); attr != nil {
		attrs = append(attrs, attr)
	}
	attrs = append(attrs, g.Text(""))
	return Span(attrs...)
}

// ActionableAvatar renders avatar chrome for the given interactive mode.
func ActionableAvatar(author AuthorInfo, mode AuthorInteractive) g.Node {
	if author.BlurAvatar {
		return moderatedAvatar(author, mode)
	}
	glyph := avatarGlyph(author.Avatar, author.Handle, author.DID, author.IsLabeler, author.DisplayName)
	if glyph == nil {
		return nil
	}
	attrs := []g.Node{
		g.Attr("class", profileIconClass(false)),
	}
	if author.IsLabeler {
		attrs = append(attrs, g.Attr("data-kind", "labeler"))
	}
	if attr := authorInteractiveAttr(mode); attr != nil {
		attrs = append(attrs, attr)
	}
	switch mode {
	case AuthorLinked, AuthorPeekable:
		linkAttrs := []g.Node{
			g.Attr("href", actor.ProfilePath(author.Handle, author.DID)),
			g.Attr("style", "pointer-events: auto"),
		}
		linkAttrs = append(linkAttrs, attrs...)
		linkAttrs = append(linkAttrs, glyph)
		return A(linkAttrs...)
	default:
		attrs = append(attrs, glyph)
		return Span(attrs...)
	}
}

// AvatarGlyph renders a small avatar image or fallback emoji for inline use.
func AvatarGlyph(avatarURL, handle, did string, isLabeler bool) g.Node {
	return avatarGlyph(avatarURL, handle, did, isLabeler, "")
}

func AuthorName(author AuthorInfo, mode AuthorInteractive) g.Node {
	return authorTextRef(author, mode, "author-name", author.DisplayName)
}

func AuthorHandle(author AuthorInfo, mode AuthorInteractive) g.Node {
	return authorTextRef(author, mode, "author-handle", "@"+author.Handle)
}

func authorTextRef(author AuthorInfo, mode AuthorInteractive, class, text string) g.Node {
	switch mode {
	case AuthorLinked, AuthorPeekable:
		attrs := []g.Node{
			g.Attr("href", actor.ProfilePath(author.Handle, author.DID)),
			g.Attr("class", class),
			g.Attr("style", "pointer-events: auto"),
		}
		if attr := authorInteractiveAttr(mode); attr != nil {
			attrs = append(attrs, attr)
		}
		attrs = append(attrs, g.Text(text))
		return A(attrs...)
	default:
		return Span(g.Attr("class", class), g.Text(text))
	}
}

// AuthorLink renders peekable display name and handle as discrete profile links.
func AuthorLink(author AuthorInfo) g.Node {
	if author.DisplayName != author.Handle {
		return Span(
			AuthorName(author, AuthorPeekable),
			AuthorHandle(author, AuthorPeekable),
		)
	}
	return AuthorHandle(author, AuthorPeekable)
}

func PostByline(author AuthorInfo, createdAt, now time.Time) g.Node {
	return postBylineContent(author, createdAt, now, true, false)
}
