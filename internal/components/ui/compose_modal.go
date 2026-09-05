package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const (
	ComposeDialogID      = "compose-dialog"
	ComposeDialogTitleID = "compose-dialog-title"
	ComposeParentSlotID  = "compose-parent-slot"
)

// ComposeModal renders the shared compose dialog with an empty parent slot.
func ComposeModal(cfg ComposeFieldConfig) g.Node {
	if cfg.TextareaID == "" {
		cfg.TextareaID = "compose-text"
	}
	cfg.FormID = ComposeFormID
	cfg.ParentInputID = ComposeParentInputID
	return NativeDialog(DialogConfig{
		ID:         ComposeDialogID,
		ExtraClass: "compose-dialog",
		Content: []g.Node{
			Header(
				H2(g.Attr("id", ComposeDialogTitleID), g.Text("New post")),
				Button(
					g.Attr("type", "submit"),
					g.Attr("form", ComposeFormID),
					g.Attr("formmethod", "dialog"),
					g.Attr("formnovalidate", ""),
					g.Attr("aria-label", "Close"),
					g.Text("×"),
				),
			),
			Div(g.Attr("id", ComposeParentSlotID)),
			ComposeField(cfg),
		},
	})
}
