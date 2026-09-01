package post

import (
	"fmt"
	"time"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ThoughtThreadPage(view feedquery.ThoughtThreadView, now time.Time, suggested []ui.AuthorInfo, accounts ui.AccountMenuView, publicBaseURL string) g.Node {
	path := actor.ThreadPath(view.RootHandle, view.RootDID, view.RootID)
	meta := page.PageMeta{
		Title:        fmt.Sprintf("Thread by @%s", view.RootHandle),
		Description:  fmt.Sprintf("Thought-thread by @%s on Twisky", view.RootHandle),
		CanonicalURL: page.AbsoluteURL(publicBaseURL, path),
		Path:         path,
	}
	items := make([]g.Node, 0, len(view.Posts))
	for _, postView := range view.Posts {
		if postView.Moderation.Filtered {
			continue
		}
		items = append(items, PostInThread(postView, now))
	}
	return page.Page(
		meta,
		suggested,
		accounts,
		Header(
			g.Attr("id", "thought-thread-header"),
			ui.BackButton("/"),
			H2(g.Text("Thread")),
		),
		PostSpine("thought-thread-list", items),
	)
}
