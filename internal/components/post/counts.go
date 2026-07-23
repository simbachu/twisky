package post

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PreviousCounts carries the like/reply/repost counts a client currently has
// displayed, as reported back on a periodic counts poll. A nil field means
// the client didn't report a value (e.g. first poll), so the corresponding
// count is always treated as changed.
type PreviousCounts struct {
	Reply  *int
	Repost *int
	Like   *int
}

type countIDs struct {
	reply     string
	repost    string
	like      string
	toggle    string
	poller    string
	announcer string
}

func newCountIDs(postID string) countIDs {
	esc := url.PathEscape(postID)
	return countIDs{
		reply:     "reply-count-" + esc,
		repost:    "repost-count-" + esc,
		like:      "like-count-" + esc,
		toggle:    "counts-live-toggle-" + esc,
		poller:    "counts-poller-" + esc,
		announcer: "counts-announcer-" + esc,
	}
}

func postHref(view feedquery.PostView) string {
	return "/" + view.AuthorHandle + "/post/" + url.PathEscape(view.ID)
}

// countsLiveToggle renders the play/pause control for live counts polling.
// Clicking it replaces itself (hx-target=this) with the opposite state and
// syncs the address bar via hx-push-url.
func countsLiveToggle(view feedquery.PostView, live bool) g.Node {
	href := postHref(view)
	id := newCountIDs(view.ID).toggle
	if live {
		return Button(
			g.Attr("type", "button"),
			g.Attr("id", id),
			g.Attr("aria-pressed", "true"),
			g.Attr("aria-label", "Pause live counts"),
			g.Attr("hx-get", href+"?counts=1&live=0"),
			g.Attr("hx-target", "this"),
			g.Attr("hx-swap", "outerHTML"),
			g.Attr("hx-push-url", href),
			ui.Icon(ui.IconPause),
		)
	}
	return Button(
		g.Attr("type", "button"),
		g.Attr("id", id),
		g.Attr("aria-pressed", "false"),
		g.Attr("aria-label", "Show live counts"),
		g.Attr("hx-get", href+"?counts=1&live=1"),
		g.Attr("hx-target", "this"),
		g.Attr("hx-swap", "outerHTML"),
		g.Attr("hx-push-url", href+"?live=1"),
		ui.Icon(ui.IconPlay),
	)
}

// countsPollerData renders the hidden data element the JS scheduler
// (post-counts-live.js) reads to drive the periodic refresh loop. It carries
// no hx-trigger itself; polling is scheduled entirely in JS so it can respect
// tab visibility and back off on failures.
func countsPollerData(view feedquery.PostView, now time.Time, live, oob bool) g.Node {
	attrs := []g.Node{
		g.Attr("id", newCountIDs(view.ID).poller),
		g.Attr("hidden", ""),
		g.Attr("data-counts-poll", ""),
		g.Attr("data-live", strconv.FormatBool(live)),
	}
	if oob {
		attrs = append(attrs, g.Attr("hx-swap-oob", "true"))
	}
	if live {
		age := now.Sub(view.CreatedAt)
		interval := countsPollInterval(age)
		attrs = append(attrs,
			g.Attr("data-href", postHref(view)+"?counts=1"),
			g.Attr("data-interval-ms", strconv.Itoa(int(interval.Milliseconds()))),
			g.Attr("data-max-interval-ms", "300000"),
			g.Attr("data-replies-href", postHref(view)+"?replies=1"),
			g.Attr("data-replies-cooldown-ms", strconv.Itoa(int(repliesFetchCooldown(age).Milliseconds()))),
			g.Attr("data-burst-interval-ms", strconv.Itoa(int(burstCountsPollInterval().Milliseconds()))),
		)
	}
	return Span(attrs...)
}

// countsAnnouncer renders the visually-hidden aria-live region used to
// accessibly announce count changes; it starts empty and is only updated via
// an out-of-band swap when something actually changes.
func countsAnnouncer(postID string) g.Node {
	return Div(
		g.Attr("id", newCountIDs(postID).announcer),
		g.Attr("class", "visually-hidden"),
		g.Attr("aria-live", "polite"),
		g.Attr("aria-atomic", "true"),
	)
}

// countChanged reports whether count's fuzzy-formatted display value differs
// from the previously-reported one. A nil previous value is always treated as
// changed (e.g. the client didn't report one).
func countChanged(previous *int, count int) bool {
	if previous == nil {
		return true
	}
	return ui.FormatFuzzyNumber(*previous) != ui.FormatFuzzyNumber(count)
}

// CountsRefreshFragment renders out-of-band updates for a post's engagement
// counts, for the periodic (no live param) counts poll. Only spans whose
// fuzzy-formatted value actually changed are included, plus an accessible
// announcer update if anything changed at all. When anything changes, the
// poller element is also refreshed so age-based reply cooldown stays current.
func CountsRefreshFragment(view feedquery.PostView, previous PreviousCounts, now time.Time) g.Node {
	ids := newCountIDs(view.ID)
	var nodes []g.Node
	var announced []string

	if countChanged(previous.Reply, view.ReplyCount) {
		nodes = append(nodes, ui.FuzzyCountSpan(ids.reply, view.ReplyCount, true))
		announced = append(announced, countAnnouncement(view.ReplyCount, "reply", "replies"))
	}
	if countChanged(previous.Repost, view.RepostCount) {
		nodes = append(nodes, ui.FuzzyCountSpan(ids.repost, view.RepostCount, true))
		announced = append(announced, countAnnouncement(view.RepostCount, "repost", "reposts"))
	}
	if countChanged(previous.Like, view.LikeCount) {
		nodes = append(nodes, ui.FuzzyCountSpan(ids.like, view.LikeCount, true))
		announced = append(announced, countAnnouncement(view.LikeCount, "like", "likes"))
	}

	if len(announced) > 0 {
		nodes = append(nodes, Div(
			g.Attr("id", ids.announcer),
			g.Attr("hx-swap-oob", "true"),
			g.Attr("class", "visually-hidden"),
			g.Attr("aria-live", "polite"),
			g.Attr("aria-atomic", "true"),
			g.Text(strings.Join(announced, ", ")),
		))
		// Only live clients poll counts; refresh cooldown while live.
		nodes = append(nodes, countsPollerData(view, now, true, true))
	}

	return g.Group(nodes)
}

func countAnnouncement(count int, singular, plural string) string {
	unit := plural
	if count == 1 {
		unit = singular
	}
	return ui.FormatFuzzyNumber(count) + " " + unit
}

// CountsToggleFragment renders the response to a play/pause click: the
// flipped toggle button as the main content (swapped via hx-target=this),
// plus fresh out-of-band spans and the poller state reflecting the new live
// value.
func CountsToggleFragment(view feedquery.PostView, now time.Time, live bool) g.Node {
	ids := newCountIDs(view.ID)
	return g.Group{
		countsLiveToggle(view, live),
		ui.FuzzyCountSpan(ids.reply, view.ReplyCount, true),
		ui.FuzzyCountSpan(ids.repost, view.RepostCount, true),
		ui.FuzzyCountSpan(ids.like, view.LikeCount, true),
		countsPollerData(view, now, live, true),
	}
}
