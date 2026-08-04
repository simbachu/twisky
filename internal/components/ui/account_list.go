package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// AccountList renders a vertical list of mini profiles.
func AccountList(accounts []AuthorInfo) g.Node {
	if len(accounts) == 0 {
		return nil
	}
	return Nav(
		g.Attr("class", "account-list"),
		g.Attr("aria-label", "Suggested accounts"),
		Menu(g.Group(g.Map(accounts, accountListItem))),
	)
}

func accountListItem(author AuthorInfo) g.Node {
	return Li(
		g.Attr("class", "account-list-item"),
		A(
			g.Attr("class", "account-list-item-overlay"),
			g.Attr("href", "/"+author.Handle),
			g.Attr("tabindex", "-1"),
			g.Attr("aria-hidden", "true"),
		),
		MiniProfile(author, AuthorPeekable),
		PillButton(ActionButtonConfig{
			Icon:  IconFollow,
			Label: "Follow " + author.DisplayName,
		}),
	)
}
