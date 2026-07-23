package post_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/richtext"
)

func TestPostPage_RendersAncestorsSlot(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "root",
			AuthorHandle:      "bsky.app",
			AuthorDisplayName: "Bluesky",
			Text:              "linked post",
		},
		HasAncestors: true,
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `id="post-page-ancestors"`) {
		t.Fatalf("html = %q, want post-page-ancestors slot", html)
	}
	if !strings.Contains(html, `hx-get="/bsky.app/post/root?ancestors=1"`) {
		t.Fatalf("html = %q, want ancestors fragment hx-get", html)
	}
	if strings.Contains(html, `id="post-parent"`) || strings.Contains(html, `id="post-grandparent"`) {
		t.Fatalf("html = %q, want no inline ancestor posts", html)
	}
	if !strings.Contains(html, `class="post post-page"`) {
		t.Fatalf("html = %q, want focus post-page article", html)
	}
	if !strings.Contains(html, `post-page-ancestors.js`) {
		t.Fatalf("html = %q, want post-page-ancestors.js script", html)
	}

	ancestorsIdx := strings.Index(html, `id="post-page-ancestors"`)
	focusIdx := strings.Index(html, `class="post post-page"`)
	if ancestorsIdx < 0 || focusIdx < 0 || ancestorsIdx > focusIdx {
		t.Fatalf("html order wrong: ancestors@%d focus@%d", ancestorsIdx, focusIdx)
	}
}

func TestPostPage_RootPostRendersLiveCountsFeatures(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:           "root",
			AuthorHandle: "bsky.app",
			Text:         "brand new post",
			CreatedAt:    time.Now().UTC(),
			LikeCount:    5,
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="like-count-root"`,
		`class="fuzzy-number"`,
		`id="counts-poller-root"`,
		`data-counts-poll`,
		`id="counts-announcer-root"`,
		`class="visually-hidden"`,
		`aria-label="Pause live counts"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestPostPage_RootPostAutoStartsLiveOnlyWhenFresh(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:           "root",
			AuthorHandle: "bsky.app",
			CreatedAt:    time.Now().Add(-time.Hour).UTC(),
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Show live counts"`) {
		t.Fatalf("html = %q, want an old post to default to paused", html)
	}
	if strings.Contains(html, `data-href`) {
		t.Fatalf("html = %q, want no scheduler data-href while paused", html)
	}
}

func TestPostPage_ExplicitLiveStartsOldPostLive(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:           "root",
			AuthorHandle: "bsky.app",
			CreatedAt:    time.Now().Add(-time.Hour).UTC(),
		},
		ExplicitLive: true,
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `aria-label="Pause live counts"`) {
		t.Fatalf("html = %q, want ?live=1 to start an old post live", html)
	}
}

func TestPostPage_OmitsAncestorsWithoutThreadContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "root",
			AuthorHandle:      "bsky.app",
			AuthorDisplayName: "Bluesky",
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `id="post-page-ancestors"`) {
		t.Fatalf("html = %q, want no ancestors slot", html)
	}
}

func TestPostPageAncestors_RendersAncestorPosts(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPageAncestors(feedquery.PostPageView{
		Ancestors: []feedquery.AncestorNodeView{
			{Post: feedquery.PostView{ID: "parent", AuthorHandle: "bsky.app", Text: "parent post"}},
			{Post: feedquery.PostView{ID: "grandparent", AuthorHandle: "other.example", Text: "grandparent post"}},
		},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "parent post") || !strings.Contains(html, "grandparent post") {
		t.Fatalf("html = %q, want ancestor post text", html)
	}
	grandparentIdx := strings.Index(html, `id="post-grandparent"`)
	parentIdx := strings.Index(html, `id="post-parent"`)
	if grandparentIdx < 0 || parentIdx < 0 || grandparentIdx > parentIdx {
		t.Fatalf("html order wrong: grandparent@%d parent@%d", grandparentIdx, parentIdx)
	}
	if strings.Contains(html, `<html`) {
		t.Fatalf("html = %q, want fragment without page wrapper", html)
	}
}

func TestPostPageAncestors_RendersUnavailableAncestor(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPageAncestors(feedquery.PostPageView{
		Ancestors: []feedquery.AncestorNodeView{
			{Unavailable: true},
		},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "Post unavailable") {
		t.Fatalf("html = %q, want unavailable ancestor message", html)
	}
}

func TestPostPageAncestors_RendersFilteredAncestor(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPageAncestors(feedquery.PostPageView{
		Ancestors: []feedquery.AncestorNodeView{
			{
				Post: feedquery.PostView{
					ID:           "parent",
					AuthorHandle: "bsky.app",
					Text:         "parent post",
					Moderation:   feedquery.ModerationView{Filtered: true},
				},
			},
		},
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "Post hidden by moderation") {
		t.Fatalf("html = %q, want filtered ancestor message", html)
	}
	if strings.Contains(html, "parent post") {
		t.Fatalf("html = %q, want filtered ancestor content hidden", html)
	}
}

func TestPostPage_RendersNestedReplyTree(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "root",
			AuthorHandle:      "bsky.app",
			AuthorDisplayName: "Bluesky",
			Text:              "root post",
		},
		Replies: []feedquery.ThreadNodeView{
			{
				Post: feedquery.PostView{
					ID:           "reply1",
					AuthorHandle: "dev.example",
					Text:         "reply one",
				},
				Replies: []feedquery.ThreadNodeView{
					{
						Post: feedquery.PostView{
							ID:           "reply2",
							AuthorHandle: "dev.example",
							Text:         "nested reply",
						},
					},
				},
			},
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="post post-page"`) {
		t.Fatalf("html = %q, want root post-page article", html)
	}

	rootArticleIdx := strings.Index(html, `class="post post-page"`)
	repliesIdx := strings.Index(html, `class="post-replies"`)
	if rootArticleIdx < 0 || repliesIdx < 0 || repliesIdx < rootArticleIdx {
		t.Fatalf("html = %q, want post-replies inside root article", html)
	}

	replyOneIdx := strings.Index(html, `id="post-reply1"`)
	nestedReplyIdx := strings.Index(html, `id="post-reply2"`)
	nestedRepliesIdx := strings.LastIndex(html, `class="post-replies"`)
	if replyOneIdx < 0 || nestedReplyIdx < 0 || nestedRepliesIdx < 0 {
		t.Fatalf("html = %q, want reply and nested reply articles", html)
	}
	if nestedRepliesIdx < replyOneIdx || nestedReplyIdx < nestedRepliesIdx {
		t.Fatalf("html = %q, want nested reply inside nested post-replies list", html)
	}

	if strings.Contains(html, `href="/dev.example/post/reply1"`) {
		t.Fatalf("html = %q, want no link wrapper around reply article", html)
	}
	if strings.Contains(html, `href="/dev.example/post/reply2"`) {
		t.Fatalf("html = %q, want no link wrapper around nested reply article", html)
	}
}

func TestPostPage_RendersSocialMetaFromPostText(t *testing.T) {
	t.Parallel()

	published := time.Date(2026, 7, 22, 10, 30, 0, 0, time.UTC)
	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "abc",
			AuthorHandle:      "bsky.app",
			AuthorDisplayName: "Bluesky",
			Text:              "hello from the feed #bluesky",
			CreatedAt:         published,
			TextSegments: []richtext.Segment{
				{Kind: richtext.Plain, Text: "hello from the feed "},
				{Kind: richtext.Tag, Text: "#bluesky", Tag: "bluesky"},
			},
		},
	}, time.Now().UTC(), nil, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:title" content="Bluesky (@bsky.app)"`,
		`property="og:description" content="hello from the feed #bluesky"`,
		`property="og:type" content="article"`,
		`property="og:url" content="https://twisky.test/bsky.app/post/abc"`,
		`property="article:published_time" content="2026-07-22T10:30:00Z"`,
		`property="article:author" content="https://twisky.test/bsky.app"`,
		`property="article:tag" content="bluesky"`,
		`name="twitter:creator" content="@bsky.app"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestPostPage_PrefersPostImageOverLinkPreview(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "abc",
			AuthorHandle:      "bsky.app",
			AuthorDisplayName: "Bluesky",
			Text:              "with media",
			Images: []feedquery.ImageView{{
				Fullsize: "https://cdn.example/post.jpg",
				Alt:      "a landscape",
				Width:    1200,
				Height:   675,
			}},
			LinkPreviewMaybe: &feedquery.LinkPreviewView{
				Thumb: "https://cdn.example/link.jpg",
			},
			AuthorAvatar: "https://cdn.example/avatar.jpg",
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:image" content="https://cdn.example/post.jpg"`,
		`name="twitter:card" content="summary_large_image"`,
		`property="og:image:width" content="1200"`,
		`property="og:image:height" content="675"`,
		`property="og:image:alt" content="a landscape"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestPostPage_UsesModerationFallbackForFilteredPost(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "abc",
			AuthorHandle:      "bsky.app",
			AuthorDisplayName: "Bluesky",
			Text:              "hidden content",
			Moderation:        feedquery.ModerationView{Filtered: true},
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:title" content="Bluesky (@bsky.app)"`,
		`property="og:description" content="Post hidden by moderation on Twisky"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, `property="og:image"`) {
		t.Fatalf("html = %q, want no og:image for filtered post", html)
	}
}

func TestPostPage_RendersReplyContextInSocialMeta(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "abc",
			AuthorHandle:      "alice.example",
			AuthorDisplayName: "Alice",
			Text:              "my take on this",
			ReplyCount:        2,
		},
		HasAncestors: true,
		ReplyParentMaybe: &feedquery.AuthorView{
			Handle:      "bob.example",
			DisplayName: "Bob",
		},
	}, time.Now().UTC(), nil, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:title" content="Reply by Alice (@alice.example)"`,
		`property="og:description" content="Replying to @bob.example · my take on this · 2 replies"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestPostPage_RendersQuoteAndLinkContextInSocialMeta(t *testing.T) {
	t.Parallel()

	quoted := feedquery.PostView{
		ID:                "quoted",
		AuthorHandle:      "carol.example",
		AuthorDisplayName: "Carol",
		Text:              "hot take",
	}
	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "abc",
			AuthorHandle:      "alice.example",
			AuthorDisplayName: "Alice",
			Text:              "short",
			QuotedPostMaybe:   &quoted,
			LinkPreviewMaybe: &feedquery.LinkPreviewView{
				URI:   "https://example.com/article",
				Title: "Example Site Title",
			},
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	want := `property="og:description" content="short · Quoting @carol.example: hot take · Example Site Title"`
	if !strings.Contains(html, want) {
		t.Fatalf("html = %q, want %s", html, want)
	}
}

func TestPostPage_UsesAvatarAsLargeImageCard(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "abc",
			AuthorHandle:      "bsky.app",
			AuthorDisplayName: "Bluesky",
			Text:              "text only",
			AuthorAvatar:      "https://cdn.example/avatar.jpg",
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:image" content="https://cdn.example/avatar.jpg"`,
		`name="twitter:card" content="summary_large_image"`,
		`property="og:image:alt" content="Bluesky (@bsky.app) on Twisky"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestPostPage_UsesQuotedImageWhenNoOwnMedia(t *testing.T) {
	t.Parallel()

	quoted := feedquery.PostView{
		ID:           "quoted",
		AuthorHandle: "carol.example",
		Images: []feedquery.ImageView{{
			Fullsize: "https://cdn.example/quoted.jpg",
		}},
	}
	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "abc",
			AuthorHandle:      "alice.example",
			AuthorDisplayName: "Alice",
			Text:              "quoting",
			AuthorAvatar:      "https://cdn.example/avatar.jpg",
			QuotedPostMaybe:   &quoted,
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `property="og:image" content="https://cdn.example/quoted.jpg"`) {
		t.Fatalf("html = %q, want quoted image in og:image", html)
	}
}

func TestPostPage_UsesThreadFallbackWhenParentUnavailable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "abc",
			AuthorHandle:      "alice.example",
			AuthorDisplayName: "Alice",
			Text:              "still here",
		},
		HasAncestors: true,
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	want := `property="og:description" content="Reply in thread · still here"`
	if !strings.Contains(html, want) {
		t.Fatalf("html = %q, want %s", html, want)
	}
	if strings.Contains(html, `property="og:title" content="Reply by`) {
		t.Fatalf("html = %q, want plain title when parent author unknown", html)
	}
}
