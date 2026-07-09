package feed_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/feed"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestFeedItems_RendersSentinelWhenNextCursorPresent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := feed.FeedItems(feedquery.FeedView{
		Posts: []feedquery.PostView{{
			ID:           "post",
			AuthorHandle: "dev.example",
			Text:         "hello",
		}},
		NextCursor: "next-page",
	}, time.Now().UTC(), "/dev.example").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="feed-sentinel"`) {
		t.Fatalf("html = %q, want feed-sentinel", html)
	}
	if !strings.Contains(html, `hx-get="/dev.example?cursor=next-page"`) {
		t.Fatalf("html = %q, want cursor hx-get", html)
	}
	if !strings.Contains(html, `hx-trigger="revealed"`) {
		t.Fatalf("html = %q, want revealed trigger", html)
	}
}

func TestFeedItems_OmitsSentinelWhenNoNextCursor(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := feed.FeedItems(feedquery.FeedView{
		Posts: []feedquery.PostView{{
			ID:           "post",
			AuthorHandle: "dev.example",
			Text:         "hello",
		}},
	}, time.Now().UTC(), "/dev.example").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `class="feed-sentinel"`) {
		t.Fatalf("html = %q, want no feed-sentinel", html)
	}
}

func TestNewPostsBanner_ReturnsNilWhenCountZero(t *testing.T) {
	t.Parallel()

	node := feed.NewPostsBanner(0, "/dev.example", "top")
	if node != nil {
		t.Fatalf("NewPostsBanner() = %v, want nil", node)
	}
}

func TestNewPostsBanner_RendersButtonWhenCountNonZero(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := feed.NewPostsBanner(3, "/dev.example", "top").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="new-posts-button"`) {
		t.Fatalf("html = %q, want new-posts-button", html)
	}
	if !strings.Contains(html, `hx-get="/dev.example?refresh=top"`) {
		t.Fatalf("html = %q, want refresh hx-get", html)
	}
	if !strings.Contains(html, `hx-target="#feed-list"`) {
		t.Fatalf("html = %q, want feed-list target", html)
	}
	if !strings.Contains(html, "Show 3 new posts") {
		t.Fatalf("html = %q, want banner label", html)
	}
}

func TestNewPostsPoll_RendersPollerAttributes(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := feed.NewPostsPoll("/dev.example", "top").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `id="new-posts-slot"`) {
		t.Fatalf("html = %q, want new-posts-slot id", html)
	}
	if !strings.Contains(html, `hx-get="/dev.example?since=top"`) {
		t.Fatalf("html = %q, want since hx-get", html)
	}
	if !strings.Contains(html, `hx-trigger="every 20s"`) {
		t.Fatalf("html = %q, want every 20s trigger", html)
	}
}

func TestNewPostsPollOOB_RendersOutOfBandSwap(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := feed.NewPostsPollOOB("/dev.example", "new-top").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `hx-swap-oob="true"`) {
		t.Fatalf("html = %q, want hx-swap-oob", html)
	}
	if !strings.Contains(html, `hx-get="/dev.example?since=new-top"`) {
		t.Fatalf("html = %q, want updated since id", html)
	}
}
