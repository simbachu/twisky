package post_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/richtext"
)

func TestPost_RendersFacetSpanClasses(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "dev.example",
		TextSegments: []richtext.Segment{
			{Kind: richtext.Mention, Text: "@dev.example", Mention: "dev.example"},
			{Kind: richtext.Plain, Text: " see "},
			{Kind: richtext.Tag, Text: "#golang", Tag: "golang"},
			{Kind: richtext.Plain, Text: " "},
			{Kind: richtext.Link, Text: "https://example.com", URI: "https://example.com"},
		},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, class := range []string{`class="facet-mention"`, `class="facet-tag"`, `class="facet-link"`} {
		if !strings.Contains(html, class) {
			t.Fatalf("html = %q, want %s", html, class)
		}
	}
}

func TestPost_RendersTimestamp(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 12, 18, 43, 28, 0, time.UTC)
	now := time.Date(2026, 3, 12, 20, 43, 28, 0, time.UTC)
	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "simbachu.com",
		CreatedAt:    createdAt,
	}, now).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `<time datetime="2026-03-12T18:43:28Z"`) {
		t.Fatalf("html = %q, want time element with datetime attribute", html)
	}
	if !strings.Contains(html, `title="Mar 12, 2026, 6:43 PM UTC"`) {
		t.Fatalf("html = %q, want absolute time in title", html)
	}
	if !strings.Contains(html, ">2h<") {
		t.Fatalf("html = %q, want relative visible timestamp", html)
	}
}

func TestPost_OmitsTimestampWhenMissing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "simbachu.com",
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	if strings.Contains(buf.String(), "<time") {
		t.Fatalf("html = %q, want no time element when CreatedAt is zero", buf.String())
	}
}

func TestPost_RendersImageCountClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		imageCount int
		wantClass string
	}{
		{name: "one image", imageCount: 1, wantClass: `class="post-images post-images-1"`},
		{name: "two images", imageCount: 2, wantClass: `class="post-images post-images-2"`},
		{name: "three images", imageCount: 3, wantClass: `class="post-images post-images-3"`},
		{name: "four images", imageCount: 4, wantClass: `class="post-images post-images-4"`},
		{name: "five images clamps to four", imageCount: 5, wantClass: `class="post-images post-images-4"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			images := make([]feedquery.ImageView, tc.imageCount)
			for i := range images {
				images[i] = feedquery.ImageView{
					Thumb: fmt.Sprintf("https://example.com/thumb-%d.jpg", i),
					Alt:   fmt.Sprintf("image %d", i),
				}
			}

			var buf bytes.Buffer
			if err := post.Post(feedquery.PostView{
				ID:           "abc123",
				AuthorHandle: "dev.example",
				Images:       images,
			}, time.Now().UTC()).Render(&buf); err != nil {
				t.Fatalf("Render() err = %v", err)
			}

			if !strings.Contains(buf.String(), tc.wantClass) {
				t.Fatalf("html = %q, want %s", buf.String(), tc.wantClass)
			}
		})
	}
}

func TestInsetPost_RendersImageCountClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		imageCount int
		wantClass string
	}{
		{name: "one image", imageCount: 1, wantClass: `class="post-images post-images-1"`},
		{name: "two images", imageCount: 2, wantClass: `class="post-images post-images-2"`},
		{name: "three images", imageCount: 3, wantClass: `class="post-images post-images-3"`},
		{name: "four images", imageCount: 4, wantClass: `class="post-images post-images-4"`},
		{name: "five images clamps to four", imageCount: 5, wantClass: `class="post-images post-images-4"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			images := make([]feedquery.ImageView, tc.imageCount)
			for i := range images {
				images[i] = feedquery.ImageView{
					Thumb: fmt.Sprintf("https://example.com/thumb-%d.jpg", i),
					Alt:   fmt.Sprintf("image %d", i),
				}
			}

			view := feedquery.PostView{
				ID:           "abc123",
				AuthorHandle: "dev.example",
				Images:       images,
			}

			var buf bytes.Buffer
			if err := post.InsetPost(&view, time.Now().UTC()).Render(&buf); err != nil {
				t.Fatalf("Render() err = %v", err)
			}

			if !strings.Contains(buf.String(), tc.wantClass) {
				t.Fatalf("html = %q, want %s", buf.String(), tc.wantClass)
			}
		})
	}
}

func TestPost_RendersQuotedPostInset(t *testing.T) {
	t.Parallel()

	quoted := feedquery.PostView{
		ID:           "quoted",
		AuthorHandle: "quoted.example",
		Text:         "original post",
	}
	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:              "qrt",
		AuthorHandle:    "dev.example",
		Text:            "my take",
		QuotedPostMaybe: &quoted,
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="post inset-post"`) {
		t.Fatalf("html = %q, want post inset-post class", html)
	}
	if !strings.Contains(html, "original post") {
		t.Fatalf("html = %q, want quoted post text", html)
	}
	if !strings.Contains(html, "my take") {
		t.Fatalf("html = %q, want main post text", html)
	}
}
