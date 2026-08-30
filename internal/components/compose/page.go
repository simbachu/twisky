package compose

import (
	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// NewPostPage renders the full-page compose surface.
func NewPostPage(errorMessage, text, publicBaseURL string, suggested []ui.AuthorInfo, accounts ui.AccountMenuView) g.Node {
	children := []g.Node{
		Header(
			ui.BackButton("/"),
			H1(g.Text("New post")),
		),
		ui.ComposeField(ui.ComposeFieldConfig{
			TextareaID: "new-post-text",
			Text:       text,
			Error:      errorMessage,
		}),
	}
	return page.Page(
		page.PageMeta{
			Title:        "New post · Twisky",
			Description:  "Compose a new post on Twisky.",
			CanonicalURL: page.AbsoluteURL(publicBaseURL, "/my/posts/new"),
			Path:         "/my/posts/new",
			OGType:       "website",
		},
		suggested,
		accounts,
		children...,
	)
}
