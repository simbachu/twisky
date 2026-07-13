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

	ATProtoRepoURL    = "https://github.com/bluesky-social/atproto"
	BlueskyAPIRepoURL = "https://github.com/bluesky-social/indigo"
	BlueskySocialURL  = "https://github.com/bluesky-social"
	TwiskyRepoURL     = "https://github.com/simbachu/twisky"
)

func externalLink(label, href string) g.Node {
	return A(
		g.Attr("href", href),
		g.Attr("target", "_blank"),
		g.Attr("rel", "noopener noreferrer"),
		g.Text(label),
	)
}

func ProtocolAttribution() g.Node {
	return P(
		g.Text("Implements the "),
		externalLink("ATProto protocol", ATProtoRepoURL),
		g.Text(" and the "),
		externalLink("Bluesky API", BlueskyAPIRepoURL),
		g.Text(" by "),
		externalLink("@bluesky.social", BlueskySocialURL),
	)
}

func VersionInfo() g.Node {
	return P(
		externalLink(AppName, TwiskyRepoURL),
		g.Text(fmt.Sprintf(" Version: %s", Version)),
	)
}

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
				Script(
					g.Attr("src", "https://cdn.jsdelivr.net/npm/hls.js@1.6.2/dist/hls.min.js"),
					g.Attr("integrity", "sha384-QHoMEQEjeievZsHu5ejPFm+o1o93XoWIEziW/+oc9LLMGsPNAbp1pN4PHhI/KIzW"),
					g.Attr("crossorigin", "anonymous"),
					g.Attr("defer", ""),
				),
				Script(g.Attr("src", "/static/scripts/post-video.js"), g.Attr("defer", "")),
				Script(g.Attr("src", "/static/scripts/post-page-ancestors.js"), g.Attr("defer", "")),
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
					ProtocolAttribution(),
					VersionInfo(),
					P(g.Text(fmt.Sprintf("© %d %s", AppCopyrightYear, AppName))),
				),
			),
		),
	)
}
