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
	}, time.Now().UTC(), nil, ui.AccountMenuView{Enabled: true}, "https://twisky.test").Render(&buf); err != nil {
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
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q; got:\n%s", want, html)
		}
	}
	if strings.Count(html, `aria-current="page"`) != 1 {
		t.Fatalf("want exactly one aria-current; got:\n%s", html)
	}
}
