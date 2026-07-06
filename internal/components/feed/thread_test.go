package feed_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/feed"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestFeedThread_RendersParentInsetAndChildPost(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := feed.FeedThread(
		feedquery.PostView{
			ID:           "parent",
			AuthorHandle: "other.example",
			Text:         "parent post",
		},
		feedquery.PostView{
			ID:           "reply",
			AuthorHandle: "dev.example",
			Text:         "a reply",
		},
	).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="feed-thread"`) {
		t.Fatalf("html = %q, want feed-thread class", html)
	}
	if !strings.Contains(html, `class="post inset-post"`) {
		t.Fatalf("html = %q, want post inset-post class", html)
	}
	if !strings.Contains(html, `class="post"`) {
		t.Fatalf("html = %q, want post class", html)
	}
	if !strings.Contains(html, "parent post") {
		t.Fatalf("html = %q, want parent post text", html)
	}
	if !strings.Contains(html, "a reply") {
		t.Fatalf("html = %q, want reply text", html)
	}
}

func TestFeed_RendersThreadWhenReplyParentPresent(t *testing.T) {
	t.Parallel()

	parent := feedquery.PostView{
		ID:           "parent",
		AuthorHandle: "other.example",
		Text:         "parent post",
	}
	var buf bytes.Buffer
	if err := feed.Feed(feedquery.FeedView{
		Posts: []feedquery.PostView{{
			ID:               "reply",
			AuthorHandle:     "dev.example",
			Text:             "a reply",
			ReplyParentMaybe: &parent,
		}},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="feed-thread"`) {
		t.Fatalf("html = %q, want feed-thread class", html)
	}
	if !strings.Contains(html, "parent post") {
		t.Fatalf("html = %q, want parent post text", html)
	}
}

func TestFeed_RendersPlainPostWithoutParent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := feed.Feed(feedquery.FeedView{
		Posts: []feedquery.PostView{{
			ID:           "post",
			AuthorHandle: "dev.example",
			Text:         "standalone post",
		}},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `class="feed-thread"`) {
		t.Fatalf("html = %q, want no feed-thread class", html)
	}
	if !strings.Contains(html, "standalone post") {
		t.Fatalf("html = %q, want standalone post text", html)
	}
}
