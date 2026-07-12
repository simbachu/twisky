package feed_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/feed"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestFeed_RendersReplyMetaInsteadOfParentInset(t *testing.T) {
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
	}, time.Now().UTC(), "/dev.example").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `class="feed-thread"`) {
		t.Fatalf("html = %q, want no feed-thread class", html)
	}
	if strings.Contains(html, `class="post inset-post"`) {
		t.Fatalf("html = %q, want no inset parent post", html)
	}
	if strings.Contains(html, "parent post") {
		t.Fatalf("html = %q, want parent text omitted from feed", html)
	}
	if !strings.Contains(html, "⤷ Reply to @other.example") {
		t.Fatalf("html = %q, want reply meta line", html)
	}
	if !strings.Contains(html, "a reply") {
		t.Fatalf("html = %q, want reply text", html)
	}
}

func TestFeed_RendersOverlayBehindContent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := feed.Feed(feedquery.FeedView{
		Posts: []feedquery.PostView{{
			ID:           "post",
			AuthorHandle: "dev.example",
			Text:         "standalone post",
		}},
	}, time.Now().UTC(), "/dev.example").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	overlayIdx := strings.Index(html, `aria-label="View post"`)
	contentIdx := strings.Index(html, `<article class="post"`)
	if overlayIdx < 0 || contentIdx < 0 {
		t.Fatalf("html = %q, want feed overlay link and post content", html)
	}
	if overlayIdx > contentIdx {
		t.Fatalf("html = %q, want overlay before content in DOM", html)
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
	}, time.Now().UTC(), "/dev.example").Render(&buf); err != nil {
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
