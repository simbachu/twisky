package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const ComposeDialogID = "compose-dialog"

// ComposeModal renders the home compose dialog with the shared field.
func ComposeModal(cfg ComposeFieldConfig) g.Node {
	if cfg.TextareaID == "" {
		cfg.TextareaID = "compose-text"
	}
	return NativeDialog(DialogConfig{
		ID:         ComposeDialogID,
		ExtraClass: "compose-dialog",
		Content: []g.Node{
			Header(
				H2(g.Text("New post")),
				Button(
					g.Attr("type", "submit"),
					g.Attr("formmethod", "dialog"),
					g.Attr("aria-label", "Close"),
					g.Text("×"),
				),
			),
			ComposeField(cfg),
		},
	})
}
