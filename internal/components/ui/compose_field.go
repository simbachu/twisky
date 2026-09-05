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
	// ParentURI, when set, is the AT URI of the post being replied to.
	ParentURI string
	// FormID, when set, is the form element id (needed for the dialog close control).
	FormID string
	// ParentInputID, when set, is the hidden parent input id.
	ParentInputID string
}

const composeFormAction = "/my/posts"

// ComposeFormID is the id of the shared compose dialog form.
const ComposeFormID = "compose-form"

// ComposeParentInputID is the id of the hidden parent field in the shared dialog.
const ComposeParentInputID = "compose-parent-uri"

// ComposeField renders the shared post textarea and submit control.
func ComposeField(cfg ComposeFieldConfig) g.Node {
	textareaID := cfg.TextareaID
	if textareaID == "" {
		textareaID = "compose-text"
	}
	parentInput := []g.Node{
		g.Attr("type", "hidden"),
		g.Attr("name", "parent"),
		g.Attr("value", cfg.ParentURI),
	}
	if cfg.ParentInputID != "" {
		parentInput = append(parentInput, g.Attr("id", cfg.ParentInputID))
	}
	children := []g.Node{
		Label(
			g.Attr("for", textareaID),
			g.Text("Post text"),
		),
		Input(g.Group(parentInput)),
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
	formAttrs := []g.Node{
		g.Attr("class", "compose-field"),
		g.Attr("method", "post"),
		g.Attr("action", composeFormAction),
	}
	if cfg.FormID != "" {
		formAttrs = append(formAttrs, g.Attr("id", cfg.FormID))
	}
	return Form(
		g.Group(formAttrs),
		g.Group(children),
	)
}
