package post

import (
	"fmt"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func opThreadNumberingLink(view feedquery.PostView) g.Node {
	index, count, ok := view.OPThreadNumber()
	if !ok {
		return nil
	}
	rootHandle, rootDID, rootID := view.ThreadRootPathParts()
	href := actor.ThreadPathWithFragment(rootHandle, rootDID, rootID, view.ID)
	return P(
		g.Attr("class", "post-op-thread-number"),
		A(
			g.Attr("href", href),
			g.Attr("aria-label", fmt.Sprintf("%d of %d", index, count)),
			g.Attr("style", "pointer-events: auto"),
			ui.Icon(ui.IconThread),
			Span(g.Text(fmt.Sprintf("%d / %d", index, count))),
		),
	)
}
