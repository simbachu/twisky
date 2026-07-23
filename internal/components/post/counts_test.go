package post_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func intPtr(n int) *int { return &n }

func TestCountsRefreshFragment_OmitsUnchangedSpans(t *testing.T) {
	t.Parallel()

	view := feedquery.PostView{ID: "root", LikeCount: 1042, RepostCount: 3, ReplyCount: 1}

	var buf bytes.Buffer
	if err := post.CountsRefreshFragment(view, post.PreviousCounts{
		Like:   intPtr(1042),
		Repost: intPtr(3),
		Reply:  intPtr(1),
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	if got := buf.String(); got != "" {
		t.Fatalf("html = %q, want empty response when nothing changed", got)
	}
}

func TestCountsRefreshFragment_IncludesOnlyChangedSpan(t *testing.T) {
	t.Parallel()

	view := feedquery.PostView{ID: "root", LikeCount: 1042, RepostCount: 3, ReplyCount: 1}

	var buf bytes.Buffer
	if err := post.CountsRefreshFragment(view, post.PreviousCounts{
		Like:   intPtr(1041),
		Repost: intPtr(3),
		Reply:  intPtr(1),
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `id="like-count-root"`) || !strings.Contains(html, "1042") {
		t.Fatalf("html = %q, want the changed like span", html)
	}
	if strings.Contains(html, `id="repost-count-root"`) || strings.Contains(html, `id="reply-count-root"`) {
		t.Fatalf("html = %q, want unchanged spans omitted", html)
	}
	if !strings.Contains(html, `id="counts-announcer-root"`) {
		t.Fatalf("html = %q, want the announcer updated when something changed", html)
	}
}

func TestCountsRefreshFragment_TreatsMissingPreviousAsChanged(t *testing.T) {
	t.Parallel()

	view := feedquery.PostView{ID: "root", LikeCount: 0, RepostCount: 0, ReplyCount: 0}

	var buf bytes.Buffer
	if err := post.CountsRefreshFragment(view, post.PreviousCounts{}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="like-count-root"`,
		`id="repost-count-root"`,
		`id="reply-count-root"`,
		`id="counts-announcer-root"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s when no previous value was reported", html, want)
		}
	}
}

func TestCountsRefreshFragment_NoChangeAcrossFuzzyBoundary(t *testing.T) {
	t.Parallel()

	// 11001 -> 11500 both fuzzy-format to "11K" (n/1000), so no swap should
	// occur even though the exact counts differ.
	view := feedquery.PostView{ID: "root", LikeCount: 11500}

	var buf bytes.Buffer
	if err := post.CountsRefreshFragment(view, post.PreviousCounts{
		Like: intPtr(11001),
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	if got := buf.String(); strings.Contains(got, `id="like-count-root"`) {
		t.Fatalf("html = %q, want no swap when the fuzzy value is unchanged", got)
	}
}

func TestCountsRefreshFragment_ChangeAcrossFuzzyBoundary(t *testing.T) {
	t.Parallel()

	// 9999 -> 10001 crosses the 10K fuzzy threshold ("9999" -> "10K").
	view := feedquery.PostView{ID: "root", LikeCount: 10001}

	var buf bytes.Buffer
	if err := post.CountsRefreshFragment(view, post.PreviousCounts{
		Like: intPtr(9999),
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `id="like-count-root"`) || !strings.Contains(html, ">10K<") {
		t.Fatalf("html = %q, want a swap crossing the fuzzy boundary", html)
	}
}

func TestCountsToggleFragment_RendersButtonSpansAndPollerState(t *testing.T) {
	t.Parallel()

	view := feedquery.PostView{
		ID:           "root",
		AuthorHandle: "bsky.app",
		LikeCount:    5,
		RepostCount:  2,
		ReplyCount:   1,
		CreatedAt:    time.Now().Add(-time.Hour),
	}

	var buf bytes.Buffer
	if err := post.CountsToggleFragment(view, time.Now(), true).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`aria-pressed="true"`,
		`aria-label="Pause live counts"`,
		`id="like-count-root"`,
		`id="repost-count-root"`,
		`id="reply-count-root"`,
		`id="counts-poller-root"`,
		`data-live="true"`,
		`hx-swap-oob="true"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestCountsToggleFragment_PausedOmitsSchedulerData(t *testing.T) {
	t.Parallel()

	view := feedquery.PostView{ID: "root", AuthorHandle: "bsky.app"}

	var buf bytes.Buffer
	if err := post.CountsToggleFragment(view, time.Now(), false).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-pressed="false"`) || !strings.Contains(html, `data-live="false"`) {
		t.Fatalf("html = %q, want a paused toggle and poller state", html)
	}
	if strings.Contains(html, "data-href") {
		t.Fatalf("html = %q, want no scheduler data-href while paused", html)
	}
}
