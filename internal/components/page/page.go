package page

import (
	"fmt"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const (
	AppName          = "Twisky"
	AppCopyrightYear = 2026
	Version          = "0.1.0"
)

func Page(title string, description string, children ...g.Node) g.Node {
	return HTML(
		Doctype(
			Head(
				TitleEl(g.Text(title)),
				Meta(
					g.Attr("name", "description"),
					g.Attr("content", description),
				),
				Link(
					g.Attr("rel", "stylesheet"),
					g.Attr("href", "/static/styles/style.css"),
				),
			),
		),
		Body(
			Header(
				H1(g.Text(AppName)),
			),
			Main(children...),
			Footer(
				Section(
					P(g.Text("Implements the ATProto protocol by @ATProto.com")), // TODO: Add a link to the ATProto account
					P(g.Text(fmt.Sprintf("%s Version: %s", AppName, Version))),
					P(g.Text(fmt.Sprintf("© %d %s", AppCopyrightYear, AppName))),
				),
			),
		),
	)
}
