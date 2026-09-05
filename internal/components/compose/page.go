package compose

import (
	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageView configures the full-page compose surface.
type PageView struct {
	Error         string
	Text          string
	ParentURI     string
	Parent        g.Node
	PublicBaseURL string
	Suggested     []ui.AuthorInfo
	Accounts      ui.AccountMenuView
}

// NewPostPage renders the full-page compose surface.
func NewPostPage(view PageView) g.Node {
	title := "New post"
	description := "Compose a new post on Twisky."
	if view.ParentURI != "" {
		title = "Reply"
		description = "Reply to a post on Twisky."
	}
	children := []g.Node{
		Header(
			ui.BackButton("/"),
			H1(g.Text(title)),
		),
	}
	if view.Parent != nil {
		children = append(children, Div(g.Attr("class", "compose-parent"), view.Parent))
	}
	children = append(children, ui.ComposeField(ui.ComposeFieldConfig{
		TextareaID: "new-post-text",
		Text:       view.Text,
		Error:      view.Error,
		ParentURI:  view.ParentURI,
	}))
	return page.Page(
		page.PageMeta{
			Title:        title + " · Twisky",
			Description:  description,
			CanonicalURL: page.AbsoluteURL(view.PublicBaseURL, "/my/posts/new"),
			Path:         "/my/posts/new",
			OGType:       "website",
		},
		view.Suggested,
		view.Accounts,
		children...,
	)
}
