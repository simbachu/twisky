package post_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestPost_RendersActionPillGroups(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "dev.example",
		Text:         "hello",
		ReplyCount:   2,
		RepostCount:  3,
		LikeCount:    5,
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`aria-label="Engagement actions"`,
		`aria-label="Reply, 2"`,
		`aria-label="Repost, 3"`,
		`aria-label="Like, 5"`,
		`aria-label="Bookmark"`,
		`aria-label="Share"`,
		`aria-label="More options"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Count(html, `role="group"`) != 4 {
		t.Fatalf("html has %d role=group, want 4", strings.Count(html, `role="group"`))
	}
	if strings.Count(html, `class="iface-segmented"`) != 4 {
		t.Fatalf("html has %d iface-segmented, want 4", strings.Count(html, `class="iface-segmented"`))
	}
}

func TestPost_RendersImageCountClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		imageCount int
		wantClass  string
	}{
		{name: "one image", imageCount: 1, wantClass: `class="post-images post-images-1"`},
		{name: "two images", imageCount: 2, wantClass: `class="post-images post-images-2"`},
		{name: "three images", imageCount: 3, wantClass: `class="post-images post-images-3"`},
		{name: "four images", imageCount: 4, wantClass: `class="post-images post-images-4"`},
		{name: "five images clamps to four", imageCount: 5, wantClass: `class="post-images post-images-4"`},
	}

	for _, variant := range []string{"post", "inset"} {
		for _, tc := range tests {
			t.Run(variant+"/"+tc.name, func(t *testing.T) {
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
				now := time.Now().UTC()

				var buf bytes.Buffer
				var err error
				switch variant {
				case "post":
					err = post.Post(view, now).Render(&buf)
				case "inset":
					err = post.InsetPost(&view, now).Render(&buf)
				}
				if err != nil {
					t.Fatalf("Render() err = %v", err)
				}

				if !strings.Contains(buf.String(), tc.wantClass) {
					t.Fatalf("html = %q, want %s", buf.String(), tc.wantClass)
				}
			})
		}
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

func TestPost_RendersModerationBlurWithReveal(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "dev.example",
		Text:         "hidden text",
		Moderation: feedquery.ModerationView{
			Blurred:   true,
			AlertText: "Adult content",
		},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `<details class="post-moderation-gate"`) {
		t.Fatalf("html = %q, want moderation blur wrapper", html)
	}
	if !strings.Contains(html, "Show anyway") {
		t.Fatalf("html = %q, want reveal control", html)
	}
	if !strings.Contains(html, "Adult content") {
		t.Fatalf("html = %q, want moderation message", html)
	}
}

func TestPost_OmitsFilteredPost(t *testing.T) {
	t.Parallel()

	node := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "dev.example",
		Text:         "hidden",
		Moderation:   feedquery.ModerationView{Filtered: true},
	}, time.Now().UTC())
	if node != nil {
		t.Fatalf("Post() = %v, want nil for filtered post", node)
	}
}

func TestPost_RendersMediaBlurWithReveal(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "dev.example",
		Text:         "visible text",
		Images: []feedquery.ImageView{{
			Thumb: "https://example.com/thumb.jpg",
			Alt:   "photo",
		}},
		Moderation: feedquery.ModerationView{
			BlurMedia: true,
			AlertText: "Suggestive content",
		},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `<details class="post-moderation-gate"`) {
		t.Fatalf("html = %q, want media moderation wrapper", html)
	}
	if !strings.Contains(html, "Show media") {
		t.Fatalf("html = %q, want media reveal control", html)
	}
	if !strings.Contains(html, "visible text") {
		t.Fatalf("html = %q, want visible post text", html)
	}
}

func TestPost_RendersDefaultVideo(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "dev.example",
		Videos: []feedquery.VideoView{{
			Playlist:     "https://video.example.com/playlist.m3u8",
			Thumbnail:    "https://video.example.com/thumb.jpg",
			Alt:          "a clip",
			Presentation: "default",
			Width:        1920,
			Height:       1080,
		}},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`class="post-video"`,
		`class="post-video-player"`,
		`class="post-video-play"`,
		`poster="https://video.example.com/thumb.jpg"`,
		`data-playlist="https://video.example.com/playlist.m3u8"`,
		`data-presentation="default"`,
		`aria-label="a clip"`,
		`width="1920"`,
		`height="1080"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, `controls`) {
		t.Fatalf("html = %q, want no controls until playback starts", html)
	}
	if strings.Contains(html, `autoplay`) {
		t.Fatalf("html = %q, want no autoplay for default presentation", html)
	}
}

func TestPost_RendersGIFVideo(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "dev.example",
		Videos: []feedquery.VideoView{{
			Playlist:     "https://video.example.com/gif.m3u8",
			Thumbnail:    "https://video.example.com/gif-thumb.jpg",
			Presentation: "gif",
		}},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`data-presentation="gif"`,
		`autoplay`,
		`loop`,
		`muted`,
		`src="https://video.example.com/gif.m3u8"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, `controls`) {
		t.Fatalf("html = %q, want no controls for gif presentation", html)
	}
	if strings.Contains(html, `class="post-video-play"`) {
		t.Fatalf("html = %q, want no play overlay for gif presentation", html)
	}
}

func TestPost_RendersLinkPreview(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "link",
		AuthorHandle: "dev.example",
		Text:         "check this out",
		LinkPreviewMaybe: &feedquery.LinkPreviewView{
			URI:         "https://example.com/page",
			Title:       "Example Page",
			Description: "A useful page",
			Thumb:       "https://example.com/thumb.jpg",
		},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`class="post-link-preview"`,
		`href="https://example.com/page"`,
		`class="post-link-preview-title"`,
		`>Example Page<`,
		`class="post-link-preview-description"`,
		`>A useful page<`,
		`class="post-link-preview-host"`,
		`>example.com<`,
		`class="post-link-preview-thumb"`,
		`src="https://example.com/thumb.jpg"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestPost_RendersLinkPreviewWithoutThumb(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "link",
		AuthorHandle: "dev.example",
		LinkPreviewMaybe: &feedquery.LinkPreviewView{
			URI:   "https://example.com",
			Title: "Example",
		},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="post-link-preview"`) {
		t.Fatalf("html = %q, want link preview card", html)
	}
	if strings.Contains(html, `class="post-link-preview-thumb"`) {
		t.Fatalf("html = %q, want no thumb image", html)
	}
}

func TestPost_RendersLinkPreviewBeforeQuotedPost(t *testing.T) {
	t.Parallel()

	quoted := feedquery.PostView{
		ID:           "quoted",
		AuthorHandle: "quoted.example",
		Text:         "original post",
	}
	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "qrt",
		AuthorHandle: "dev.example",
		Text:         "my take",
		Images: []feedquery.ImageView{{
			Thumb: "https://example.com/thumb.jpg",
			Alt:   "photo",
		}},
		LinkPreviewMaybe: &feedquery.LinkPreviewView{
			URI:   "https://example.com",
			Title: "Example",
		},
		QuotedPostMaybe: &quoted,
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	imagesIdx := strings.Index(html, "post-images")
	previewIdx := strings.Index(html, "post-link-preview")
	insetIdx := strings.Index(html, "inset-post")
	if imagesIdx < 0 || previewIdx < 0 || insetIdx < 0 {
		t.Fatalf("html = %q, want images, link preview, and inset post", html)
	}
	if imagesIdx > previewIdx || previewIdx > insetIdx {
		t.Fatalf("html order wrong: images@%d preview@%d inset@%d", imagesIdx, previewIdx, insetIdx)
	}
}

func TestPost_RendersVideoMediaBlurWithReveal(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.Post(feedquery.PostView{
		ID:           "abc123",
		AuthorHandle: "dev.example",
		Text:         "visible text",
		Videos: []feedquery.VideoView{{
			Playlist:     "https://video.example.com/playlist.m3u8",
			Thumbnail:    "https://video.example.com/thumb.jpg",
			Alt:          "clip",
			Presentation: "default",
		}},
		Moderation: feedquery.ModerationView{
			BlurMedia: true,
			AlertText: "Suggestive content",
		},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `<details class="post-moderation-gate"`) {
		t.Fatalf("html = %q, want video media moderation wrapper", html)
	}
	if !strings.Contains(html, "Show media") {
		t.Fatalf("html = %q, want media reveal control", html)
	}
	if !strings.Contains(html, `class="post-video"`) {
		t.Fatalf("html = %q, want post-video inside gate", html)
	}
}
