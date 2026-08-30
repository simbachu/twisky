package login

import (
	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Page renders the chrome-less Bluesky OAuth login form.
// canonicalPath is the path used for the canonical URL (e.g. "/" or "/oauth/login").
// addingAccount is true when the visitor already has a session and is signing in another account.
func Page(errorMessage, publicBaseURL, canonicalPath string, addingAccount bool) g.Node {
	heading := "Login with Bluesky"
	lead := "Sign in with your Bluesky handle or DID."
	title := "Login with Bluesky · Twisky"
	description := "Sign in to Twisky with Bluesky OAuth."
	if addingAccount {
		heading = "Add another account"
		lead = "Sign in with another Bluesky handle or DID."
		title = "Add another account · Twisky"
		description = "Add another Bluesky account to Twisky."
	}
	children := []g.Node{
		H1(
			A(
				g.Attr("href", "/"),
				g.Attr("aria-label", page.AppName),
				Figure(ui.Icon(ui.IconBrand)),
			),
		),
		H2(g.Text(heading)),
		P(g.Text(lead)),
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
			Button(g.Attr("type", "submit"), g.Text("Login with Bluesky")),
		),
	)
	return page.Bare(
		page.PageMeta{
			Title:        title,
			Description:  description,
			CanonicalURL: page.AbsoluteURL(publicBaseURL, canonicalPath),
		},
		children...,
	)
}
