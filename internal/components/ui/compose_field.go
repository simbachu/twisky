package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ComposeFieldConfig configures the shared post compose form.
type ComposeFieldConfig struct {
	TextareaID string
	Text       string
	Error      string
}

const composeFormAction = "/my/posts"

// ComposeField renders the shared post textarea and submit control.
func ComposeField(cfg ComposeFieldConfig) g.Node {
	textareaID := cfg.TextareaID
	if textareaID == "" {
		textareaID = "compose-text"
	}
	children := []g.Node{
		Label(
			g.Attr("for", textareaID),
			g.Text("Post text"),
		),
	}
	if cfg.Error != "" {
		children = append(children, P(g.Attr("role", "alert"), g.Text(cfg.Error)))
	}
	children = append(children,
		Textarea(
			g.Attr("id", textareaID),
			g.Attr("name", "text"),
			g.Attr("rows", "4"),
			g.Attr("maxlength", "2000"),
			g.Attr("placeholder", "What's on your mind?"),
			g.Attr("required", ""),
			g.Text(cfg.Text),
		),
		Button(g.Attr("type", "submit"), g.Text("Post")),
	)
	return Form(
		g.Attr("class", "compose-field"),
		g.Attr("method", "post"),
		g.Attr("action", composeFormAction),
		g.Group(children),
	)
}
