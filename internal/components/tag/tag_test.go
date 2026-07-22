package tag_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/tag"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	tagquery "github.com/simbachu/twisky/internal/query/tag"
)

func TestTag_RendersSocialMetaTags(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := tag.Tag(tagquery.TagView{
		Tag: "art",
	}, time.Now().UTC(), nil, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:title" content="#art on Twisky"`,
		`property="og:description" content="Recent posts tagged with #art on Twisky"`,
		`property="og:type" content="website"`,
		`rel="canonical" href="https://twisky.test/tagged/art"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, `property="og:image"`) {
		t.Fatalf("html = %q, want no og:image for tag page", html)
	}
}

func TestTag_UsesFirstVisiblePostImage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := tag.Tag(tagquery.TagView{
		Tag: "art",
		Feed: feedquery.FeedView{
			Posts: []feedquery.PostView{
				{
					ID:                "hidden",
					AuthorHandle:      "a.example",
					AuthorDisplayName: "A",
					Images:            []feedquery.ImageView{{Fullsize: "https://cdn.example/hidden.jpg"}},
					Moderation:        feedquery.ModerationView{Filtered: true},
				},
				{
					ID:                "visible",
					AuthorHandle:      "b.example",
					AuthorDisplayName: "B",
					Images: []feedquery.ImageView{{
						Fullsize: "https://cdn.example/art.jpg",
						Width:    800,
						Height:   600,
						Alt:      "artwork",
					}},
				},
			},
		},
	}, time.Now().UTC(), nil, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:image" content="https://cdn.example/art.jpg"`,
		`name="twitter:card" content="summary_large_image"`,
		`property="og:image:width" content="800"`,
		`property="og:image:height" content="600"`,
		`property="og:image:alt" content="artwork"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}
