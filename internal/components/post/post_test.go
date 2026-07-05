package post_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestPost_RendersTimestamp(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 12, 20, 43, 28, 0, time.UTC)
	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "simbachu.com",
		CreatedAt:    createdAt,
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `<time datetime="2026-03-12T20:43:28Z">`) {
		t.Fatalf("html = %q, want time element with datetime attribute", html)
	}
	if !strings.Contains(html, "Mar 12, 2026, 8:43 PM UTC") {
		t.Fatalf("html = %q, want human-readable timestamp", html)
	}
}

func TestPost_OmitsTimestampWhenMissing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "simbachu.com",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	if strings.Contains(buf.String(), "<time") {
		t.Fatalf("html = %q, want no time element when CreatedAt is zero", buf.String())
	}
}
