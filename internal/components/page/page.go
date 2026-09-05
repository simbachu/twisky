package page

import (
	"fmt"

	"github.com/simbachu/twisky/internal/components/ui"
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

func protocolAttribution() g.Node {
	return P(
		Small(
			g.Text("Implements the "),
			externalLink("ATProto protocol", ATProtoRepoURL),
			g.Text(" and the "),
			externalLink("Bluesky API", BlueskyAPIRepoURL),
			g.Text(" by "),
			externalLink("@bluesky.social", BlueskySocialURL),
		),
	)
}

func versionInfo() g.Node {
	return P(
		externalLink(AppName, TwiskyRepoURL),
		g.Text(fmt.Sprintf(" Version: %s", Version)),
	)
}

func pageHeader(currentPath string, accounts ui.AccountMenuView) g.Node {
	return Header(
		H1(
			A(
				g.Attr("href", "/"),
				g.Attr("aria-label", AppName),
				Figure(ui.Icon(ui.IconBrand)),
			),
		),
		Div(
			ui.SiteNav(currentPath),
			ui.AccountMenu(accounts),
		),
	)
}

func pageFooter(suggested []ui.AuthorInfo) g.Node {
	return Footer(
		Aside(
			ui.SearchBar(),
		),
		Aside(
			Header(H3(g.Text("You might like:"))),
			ui.AccountList(suggested),
		),
		Aside(
			protocolAttribution(),
			versionInfo(),
			P(g.Text(fmt.Sprintf("© %d %s", AppCopyrightYear, AppName))),
		),
	)
}

func pageHead(meta PageMeta) []g.Node {
	headNodes := []g.Node{
		Meta(
			g.Attr("name", "viewport"),
			g.Attr("content", "width=device-width, initial-scale=1, viewport-fit=cover"),
		),
		TitleEl(g.Text(meta.Title)),
		Meta(
			g.Attr("name", "description"),
			g.Attr("content", meta.Description),
		),
	}
	headNodes = append(headNodes, socialMetaNodes(meta)...)
	return append(headNodes,
		Script(g.Raw(`(function(){var m=document.cookie.match(/(?:^|; )twisky-reply-view=([^;]*)/);if(!m)return;var v=decodeURIComponent(m[1]);if(v==="linear")document.documentElement.dataset.replyView="linear";})();`)),
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
		Script(g.Attr("src", "/static/scripts/post-counts-live.js"), g.Attr("defer", "")),
		Script(g.Attr("src", "/static/scripts/post-page-reply-view.js"), g.Attr("defer", "")),
		Script(g.Attr("src", "/static/scripts/post-share.js"), g.Attr("defer", "")),
		Script(g.Attr("src", "/static/scripts/compose-dialog.js"), g.Attr("defer", "")),
		Script(g.Attr("src", "/static/scripts/nav-back.js"), g.Attr("defer", "")),
		Script(g.Attr("src", "/static/scripts/page-alert.js"), g.Attr("defer", "")),
		Link(
			g.Attr("id", "page-favicon"),
			g.Attr("rel", "icon"),
			g.Attr("type", "image/png"),
			g.Attr("href", "/static/icons/favicon.png"),
		),
		Script(g.Attr("src", "/static/scripts/favicon-notify.js")),
	)
}

func Page(meta PageMeta, suggested []ui.AuthorInfo, accounts ui.AccountMenuView, children ...g.Node) g.Node {
	return document(meta, suggested, accounts, Div(g.Attr("id", "page-alert")), children...)
}

// PageWithAlert renders a full document with content already placed in the
// shared #page-alert region (used for query error pages).
func PageWithAlert(meta PageMeta, suggested []ui.AuthorInfo, accounts ui.AccountMenuView, alert g.Node, children ...g.Node) g.Node {
	return document(meta, suggested, accounts, Div(g.Attr("id", "page-alert"), alert), children...)
}

func document(meta PageMeta, suggested []ui.AuthorInfo, accounts ui.AccountMenuView, alertRegion g.Node, children ...g.Node) g.Node {
	mainChildren := append([]g.Node{alertRegion}, children...)
	bodyChildren := []g.Node{
		pageHeader(meta.Path, accounts),
		Main(mainChildren...),
		pageFooter(suggested),
	}
	if accounts.Current != nil {
		bodyChildren = append(bodyChildren, ui.ComposeModal(ui.ComposeFieldConfig{TextareaID: "compose-text"}))
	}
	return HTML(
		Doctype(
			Head(g.Group(pageHead(meta))),
		),
		Body(g.Group(bodyChildren)),
	)
}

// Bare renders a full document without site chrome (no sidebars / tab bar).
func Bare(meta PageMeta, children ...g.Node) g.Node {
	return HTML(
		Doctype(
			Head(g.Group(pageHead(meta))),
		),
		Body(
			Main(children...),
		),
	)
}

// PreviewPage renders a minimal HTML document for link unfurling (e.g. /healthz).
func PreviewPage(meta PageMeta, body string) g.Node {
	headNodes := []g.Node{
		TitleEl(g.Text(meta.Title)),
		Meta(
			g.Attr("name", "description"),
			g.Attr("content", meta.Description),
		),
	}
	headNodes = append(headNodes, socialMetaNodes(meta)...)

	return HTML(
		Doctype(
			Head(g.Group(headNodes)),
		),
		Body(g.Text(body)),
	)
}
