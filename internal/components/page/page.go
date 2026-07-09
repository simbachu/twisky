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
				Script(
					g.Attr("src", "https://cdn.jsdelivr.net/npm/htmx.org@2.0.10/dist/htmx.min.js"),
					g.Attr("integrity", "sha384-H5SrcfygHmAuTDZphMHqBJLc3FhssKjG7w/CeCpFReSfwBWDTKpkzPP8c+cLsK+V"),
					g.Attr("crossorigin", "anonymous"),
				),
				Link(
					g.Attr("id", "page-favicon"),
					g.Attr("rel", "icon"),
					g.Attr("type", "image/png"),
					g.Attr("href", "/static/icons/favicon.png"),
				),
				Script(g.Attr("src", "/static/scripts/favicon-notify.js")),
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
