package login

import (
	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Page(errorMessage string, suggested []ui.AuthorInfo, accounts ui.AccountMenuView, publicBaseURL string) g.Node {
	children := []g.Node{
		Header(H1(g.Text("Log in"))),
		P(g.Text("Sign in with your Bluesky handle or DID.")),
	}
	if errorMessage != "" {
		children = append(children, P(g.Attr("role", "alert"), g.Text(errorMessage)))
	}
	children = append(children,
		Form(
			g.Attr("method", "post"),
			g.Attr("action", "/oauth/login"),
			Label(
				g.Attr("for", "username"),
				g.Text("Handle or DID"),
			),
			Input(
				g.Attr("id", "username"),
				g.Attr("name", "username"),
				g.Attr("type", "text"),
				g.Attr("autocomplete", "username"),
				g.Attr("required", ""),
				g.Attr("placeholder", "you.bsky.social"),
			),
			Button(g.Attr("type", "submit"), g.Text("Continue")),
		),
	)
	return page.Page(
		page.PageMeta{
			Title:        "Log in · Twisky",
			Description:  "Sign in to Twisky with Bluesky OAuth.",
			CanonicalURL: page.AbsoluteURL(publicBaseURL, "/oauth/login"),
		},
		suggested,
		accounts,
		children...,
	)
}
