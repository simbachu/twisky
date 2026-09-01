package post_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestThoughtThreadPage_RendersPostSpine(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.ThoughtThreadPage(feedquery.ThoughtThreadView{
		RootHandle: "alice.test",
		RootDID:    "did:plc:alice",
		RootID:     "p1",
		Posts: []feedquery.PostView{
			{ID: "p1", AuthorHandle: "alice.test", Text: "first thought"},
			{ID: "p2", AuthorHandle: "alice.test", Text: "second thought"},
		},
	}, time.Now().UTC(), nil, ui.AccountMenuView{}, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="thought-thread-list"`,
		`class="post-spine"`,
		`first thought`,
		`second thought`,
		`id="post-p1"`,
		`id="post-p2"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, `class="post-op-thread-number"`) {
		t.Fatalf("html = %q, want no op-thread numbering on thread page", html)
	}
}
