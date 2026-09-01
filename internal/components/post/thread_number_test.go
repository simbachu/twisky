package post_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestPost_OPThreadNumberingLink(t *testing.T) {
	t.Parallel()

	view := feedquery.PostViewWithThreadRoot(
		feedquery.PostViewWithOPThreadNumber(feedquery.PostView{
			ID:           "p2",
			AuthorHandle: "alice.test",
			Text:         "second thought",
		}, 2, 5),
		"alice.test", "did:plc:alice", "p1",
	)

	var buf bytes.Buffer
	if err := post.Post(view, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	for _, want := range []string{
		`class="post-op-thread-number"`,
		`href="/alice.test/thread/p1#post-p2"`,
		`aria-label="2 of 5"`,
		`icon-thread-outline`,
		`2 / 5`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestPostInThread_OmitsOPThreadNumbering(t *testing.T) {
	t.Parallel()

	view := feedquery.PostViewWithThreadRoot(
		feedquery.PostViewWithOPThreadNumber(feedquery.PostView{
			ID:           "p2",
			AuthorHandle: "alice.test",
			Text:         "second thought",
		}, 2, 5),
		"alice.test", "did:plc:alice", "p1",
	)

	var buf bytes.Buffer
	if err := post.PostInThread(view, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	html := buf.String()
	if strings.Contains(html, `class="post-op-thread-number"`) {
		t.Fatalf("html = %q, want no op-thread numbering on thread page", html)
	}
}
