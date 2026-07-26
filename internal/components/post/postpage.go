package post

import (
	"net/url"
	"time"

	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ReplyView string

const (
	ReplyViewThreaded ReplyView = "threaded"
	ReplyViewLinear   ReplyView = "linear"
)

type ReplySortOrder string

const (
	ReplySortOrderHot   ReplySortOrder = "hot"
	ReplySortOrderNew   ReplySortOrder = "new"
	ReplySortOrderOld   ReplySortOrder = "old"
	ReplySortOrderRatio ReplySortOrder = "ratio" // More replies than likes ratio, "Controversial"
)

const (
	ReplyViewThreadedIcon   = "↪️"
	ReplyViewLinearIcon     = "⬇️"
	ReplySortOrderHotIcon   = "🔥"
	ReplySortOrderNewIcon   = "🆕"
	ReplySortOrderOldIcon   = "⏮️"
	ReplySortOrderRatioIcon = "🚮"
)

type PostPageSettings struct {
	ReplyView      ReplyView      `default:"threaded"`
	ReplySortOrder ReplySortOrder `default:"hot"`
	Live           bool           `default:"false"`
}

type postPageSettingOption struct {
	value string
	icon  string
	label string
}

var replyViewOptions = []postPageSettingOption{
	{string(ReplyViewThreaded), ReplyViewThreadedIcon, "Threaded"},
	{string(ReplyViewLinear), ReplyViewLinearIcon, "Linear"},
}

var replySortOptions = []postPageSettingOption{
	{string(ReplySortOrderHot), ReplySortOrderHotIcon, "Hot"},
	{string(ReplySortOrderNew), ReplySortOrderNewIcon, "New"},
	{string(ReplySortOrderOld), ReplySortOrderOldIcon, "Old"},
	{string(ReplySortOrderRatio), ReplySortOrderRatioIcon, "Controversial"},
}

func postPageSettingRadio(name, inputID string, checked bool, opt postPageSettingOption) g.Node {
	return Label(
		g.Attr("for", inputID),
		g.Attr("title", opt.label),
		Input(
			g.Attr("type", "radio"),
			g.Attr("id", inputID),
			g.Attr("name", name),
			g.Attr("value", opt.value),
			g.If(checked, g.Attr("checked", "")),
		),
		Span(g.Text(opt.icon)),
	)
}

func postPageSettingGroup(name, ariaLabel string, current string, options []postPageSettingOption) g.Node {
	return Menu(
		g.Attr("class", "iface-segmented"),
		g.Attr("role", "group"),
		g.Attr("aria-label", ariaLabel),
		g.Group(g.Map(options, func(opt postPageSettingOption) g.Node {
			return Li(postPageSettingRadio(name, name+"-"+opt.value, current == opt.value, opt))
		})),
	)
}

func postPageHeader() g.Node {
	return Header(
		g.Attr("id", "post-page-header"),
		g.Attr("tabindex", "-1"),
		H2(g.Text("Viewing post")),
	)
}

func repliesSettingsDetails(settings PostPageSettings) g.Node {
	return Details(
		Summary(g.Text("⚙"), g.Attr("aria-label", "Reply display settings")),
		Nav(
			// Reply view checked state is client-side (cookie); see post-page-reply-view.js.
			postPageSettingGroup("reply-view", "Threading mode", "", replyViewOptions),
			postPageSettingGroup("reply-sort-order", "Sort order", string(settings.ReplySortOrder), replySortOptions),
		),
	)
}

func repliesSection(replies []feedquery.ThreadNodeView, settings PostPageSettings, now time.Time, postID, opAuthorDID string, oob bool) g.Node {
	return Section(
		g.Attr("id", "post-replies"),
		g.Attr("class", "post-replies-section"),
		Header(
			H3(g.Text("Replies")),
			repliesSettingsDetails(settings),
		),
		repliesList(replies, now, repliesRootID(postID), opAuthorDID, oob),
	)
}

func PostPage(view feedquery.PostPageView, now time.Time, suggested []ui.AuthorInfo, publicBaseURL string) g.Node {
	// ReplyView is loaded from the user profile once auth exists; until then the
	// client mirrors the same threaded|linear values via cookie.
	settings := PostPageSettings{
		ReplyView:      ReplyViewThreaded,
		ReplySortOrder: ReplySortOrderHot,
		Live:           false,
	}
	return page.Page(
		postPageMeta(view, publicBaseURL),
		suggested,
		g.Group{
			g.If(view.HasAncestors, postPageAncestorsSlot(view.Post)),
			postPageHeader(),
			postPageRoot(view.Post, view.Replies, settings, now, view.ExplicitLive, feedquery.ThreadRootAuthorDID(view)),
		},
	)
}

func PostPageAncestors(view feedquery.PostPageView, now time.Time) g.Node {
	return postPageAncestorsContent(view.Ancestors, now)
}

// RepliesRefreshFragment renders an out-of-band replacement of the top-level
// replies list when the thread contains posts the client does not yet know.
// Returns an empty group when every available reply ID is already in known.
func RepliesRefreshFragment(view feedquery.PostPageView, known map[string]bool, now time.Time) g.Node {
	if !feedquery.ThreadHasUnknown(view.Replies, known) {
		return g.Group{}
	}
	return repliesList(view.Replies, now, repliesRootID(view.Post.ID), feedquery.ThreadRootAuthorDID(view), true)
}

func postPageAncestorsSlot(post feedquery.PostView) g.Node {
	href := "/" + post.AuthorHandle + "/post/" + url.PathEscape(post.ID)
	return Section(
		g.Attr("id", "post-page-ancestors"),
		g.Attr("class", "post-ancestors-section"),
		g.Attr("aria-label", "Thread context"),
		g.Attr("hx-get", href+"?ancestors=1"),
		g.Attr("hx-trigger", "twiskyAncestors"),
		g.Attr("hx-swap", "innerHTML"),
	)
}

func postPageAncestorsContent(ancestors []feedquery.AncestorNodeView, now time.Time) g.Node {
	return g.Group(g.Map(ancestors, func(ancestor feedquery.AncestorNodeView) g.Node {
		return ancestorItem(ancestor, now)
	}))
}

func ancestorItem(node feedquery.AncestorNodeView, now time.Time) g.Node {
	if node.Unavailable {
		return P(g.Text("Post unavailable"))
	}
	if node.Post.Moderation.Filtered {
		return P(g.Text(filteredPostMessage(node.Post.Moderation)))
	}
	return ClickablePostItem(Post(node.Post, now), node.Post)
}

func postPageRoot(view feedquery.PostView, replies []feedquery.ThreadNodeView, settings PostPageSettings, now time.Time, explicitLive bool, threadRootAuthorDID string) g.Node {
	if view.Moderation.Filtered {
		return P(g.Text(filteredPostMessage(view.Moderation)))
	}
	live := explicitLive || autoStartLive(now.Sub(view.CreatedAt))
	// Always render a stable replies section so live refresh can OOB-swap into
	// the list even when the page first loaded with zero replies. CSS hides
	// the section until the root ul contains li elements.
	extra := []g.Node{repliesSection(replies, settings, now, view.ID, threadRootAuthorDID, false)}
	return PostArticle(view, now, "post post-page", true, live, "", extra...)
}

func repliesRootID(postID string) string {
	return "post-replies-" + url.PathEscape(postID)
}

// repliesList renders a reply tree. rootID is set only on the top-level list
// (swap target for live refresh); nested lists pass "". When oob is true the
// list carries hx-swap-oob for out-of-band replacement.
func repliesList(replies []feedquery.ThreadNodeView, now time.Time, rootID, opAuthorDID string, oob bool) g.Node {
	attrs := []g.Node{g.Attr("class", "post-replies")}
	if rootID != "" {
		attrs = append(attrs, g.Attr("id", rootID))
	}
	if oob {
		attrs = append(attrs, g.Attr("hx-swap-oob", "true"))
	}
	return Ul(
		g.Group(attrs),
		g.Group(g.Map(replies, func(node feedquery.ThreadNodeView) g.Node {
			return replyItem(node, now, opAuthorDID)
		})),
	)
}

func replyItem(node feedquery.ThreadNodeView, now time.Time, opAuthorDID string) g.Node {
	if node.Unavailable {
		return Li(P(g.Text("Post unavailable")))
	}
	if node.Post.Moderation.Filtered {
		return Li(P(g.Text(filteredPostMessage(node.Post.Moderation))))
	}

	return Li(
		ClickablePostItem(postReply(node.Post, now, opAuthorDID), node.Post),
		g.If(len(node.Replies) > 0, repliesList(node.Replies, now, "", opAuthorDID, false)),
	)
}

func filteredPostMessage(mod feedquery.ModerationView) string {
	if mod.FilterText != "" {
		return mod.FilterText
	}
	return "Post hidden by moderation"
}
