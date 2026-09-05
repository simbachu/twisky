package home_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/home"
	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	homequery "github.com/simbachu/twisky/internal/query/home"
)

func TestHome_MarksSiteNavCurrentAndRendersFeed(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := home.Home(homequery.HomeView{
		Feed: feedquery.FeedView{
			Posts: []feedquery.PostView{{
				ID:                "abc",
				AuthorHandle:      "dev.example",
				AuthorDisplayName: "Developer",
				Text:              "hello home",
			}},
		},
	}, time.Now().UTC(), nil, ui.AccountMenuView{
		Enabled: true,
		Current: &ui.AuthorInfo{Handle: "alice.test", DID: "did:plc:alice"},
	}, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`<a href="/" aria-current="page">`,
		"hello home",
		`id="feed-list"`,
		`aria-label="Site"`,
		`property="og:title" content="Home · Twisky"`,
		`rel="canonical" href="https://twisky.test/"`,
		`href="/my/posts/new"`,
		`data-compose-open`,
		`id="compose-dialog"`,
		`id="compose-text"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q; got:\n%s", want, html)
		}
	}
	if strings.Count(html, `aria-current="page"`) != 1 {
		t.Fatalf("want exactly one aria-current; got:\n%s", html)
	}
}

func TestHome_RendersSelectedSavedFeedTabsAndMetadata(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := home.Home(homequery.HomeView{
		Feed:  feedquery.FeedView{},
		Title: "For You",
		Path:  "/feed/for-you",
		Tabs: []homequery.FeedTab{
			{Label: "Following", Href: "/", Current: false},
			{Label: "For You", Slug: "for-you", Href: "/feed/for-you", Current: true},
		},
	}, time.Now().UTC(), nil, ui.AccountMenuView{Enabled: true}, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`aria-label="Home feeds"`,
		`<a role="tab" href="/" aria-selected="false">Following</a>`,
		`<a role="tab" href="/feed/for-you" aria-selected="true" aria-current="page">For You</a>`,
		`property="og:title" content="For You · Twisky"`,
		`rel="canonical" href="https://twisky.test/feed/for-you"`,
		`<a href="/" aria-current="page">`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q; got:\n%s", want, html)
		}
	}
}
