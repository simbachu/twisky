package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MiniProfile renders the profile link for a compact account list row.
func MiniProfile(author AuthorInfo) g.Node {
	return miniProfileLink(author)
}

func miniProfileLink(author AuthorInfo) g.Node {
	children := []g.Node{
		g.Attr("href", "/"+author.Handle),
		g.Attr("class", "mini-profile"),
	}
	if author.Avatar != "" {
		children = append(children, Span(
			g.Attr("class", "byline-avatar"),
			Img(
				g.Attr("src", author.Avatar),
				g.Attr("alt", author.DisplayName),
			),
		))
	}
	children = append(children, Span(
		g.Attr("class", "mini-profile-info"),
		AuthorName(author),
		AuthorHandle(author),
	))
	return A(children...)
}

// AccountList renders a vertical list of mini profiles.
func AccountList(accounts []AuthorInfo) g.Node {
	if len(accounts) == 0 {
		return nil
	}
	return Nav(
		g.Attr("class", "account-list"),
		g.Attr("aria-label", "Suggested accounts"),
		Ul(g.Group(g.Map(accounts, accountListItem))),
	)
}

func accountListItem(author AuthorInfo) g.Node {
	return Li(
		g.Attr("class", "account-list-item"),
		miniProfileLink(author),
		PillButton(ActionButtonConfig{
			Icon:  IconFollow,
			Label: "Follow " + author.DisplayName,
		}),
	)
}
