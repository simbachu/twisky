package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// AccountMenuView is the render-facing session state for the site account menu.
type AccountMenuView struct {
	Enabled    bool
	Current    *AuthorInfo
	Additional []AuthorInfo
}

// AccountMenu renders login chrome or a disclosure for the current account.
func AccountMenu(view AccountMenuView) g.Node {
	if !view.Enabled {
		return nil
	}
	if view.Current == nil {
		return Nav(
			g.Attr("aria-label", "Account"),
			A(g.Attr("href", "/oauth/login"), g.Text("Log in")),
		)
	}

	content := []g.Node{
		Ul(
			Li(
				Form(
					g.Attr("method", "post"),
					g.Attr("action", "/oauth/logout"),
					Button(g.Attr("type", "submit"), g.Text("Log out")),
				),
			),
		),
	}
	if len(view.Additional) > 0 {
		content = append(content, accountSwitcher(view.Additional))
	}
	return Disclosure(
		"account-menu",
		MiniProfile(*view.Current, AuthorStatic),
		content...,
	)
}

func accountSwitcher(accounts []AuthorInfo) g.Node {
	return Nav(
		g.Attr("aria-label", "Switch account"),
		Header(H3(g.Text("Switch account"))),
		Menu(g.Group(g.Map(accounts, accountSwitcherItem))),
	)
}

func accountSwitcherItem(account AuthorInfo) g.Node {
	return Li(MiniProfile(account, AuthorStatic))
}
