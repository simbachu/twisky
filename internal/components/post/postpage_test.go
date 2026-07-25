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

func TestPostPage_RendersDefaultSettingsControls(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:           "root",
			AuthorHandle: "bsky.app",
			Text:         "hello",
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="post-page-header"`,
		`class="post-page-settings"`,
		`<details`,
		`<summary`,
		`Reply settings`,
		`name="reply-view"`,
		`name="reply-sort-order"`,
		`role="group"`,
		`aria-label="Reply view"`,
		`aria-label="Reply sort order"`,
		`value="threaded"`,
		`value="linear"`,
		`value="hot"`,
		`value="new"`,
		`value="old"`,
		`value="ratio"`,
		`🔥`,
		`↪️`,
		`post-page-reply-view.js`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}

	headerStart := strings.Index(html, `id="post-page-header"`)
	if headerStart < 0 {
		t.Fatalf("html = %q, want post-page-header", html)
	}
	headerEnd := strings.Index(html[headerStart:], "</header>")
	if headerEnd < 0 {
		t.Fatalf("html = %q, want closing header tag", html)
	}
	header := html[headerStart : headerStart+headerEnd]

	if !strings.Contains(header, `<details class="post-page-settings"`) {
		t.Fatalf("header = %q, want post-page-settings on details", header)
	}
	if strings.Contains(header, `<nav class="post-page-settings"`) {
		t.Fatalf("header = %q, want post-page-settings on details not nav", header)
	}

	if strings.Count(header, `name="reply-view"`) != 2 {
		t.Fatalf("header = %q, want two reply-view radios", header)
	}
	if strings.Count(header, `name="reply-sort-order"`) != 4 {
		t.Fatalf("header = %q, want four reply-sort-order radios", header)
	}
	if strings.Count(header, `checked=""`) != 1 {
		t.Fatalf("header = %q, want exactly one checked radio (sort order only)", header)
	}
	if !strings.Contains(header, `name="reply-view"`) || !strings.Contains(header, `value="threaded"`) {
		t.Fatalf("header = %q, want threaded reply-view option", header)
	}
	if !strings.Contains(header, `id="reply-view-threaded"`) {
		t.Fatalf("header = %q, want reply-view-threaded input id", header)
	}
	if strings.Contains(header, `id="reply-view-threaded"`) {
		threadedIdx := strings.Index(header, `id="reply-view-threaded"`)
		threadedSlice := header[threadedIdx:]
		if strings.Contains(threadedSlice[:strings.Index(threadedSlice, ">")], `checked=""`) {
			t.Fatalf("header = %q, want reply-view unchecked server-side (client cookie owns state)", header)
		}
	}
	hotIdx := strings.Index(header, `id="reply-sort-order-hot"`)
	if hotIdx < 0 {
		t.Fatalf("header = %q, want reply-sort-order-hot input id", header)
	}
	hotSlice := header[hotIdx:]
	if !strings.Contains(hotSlice[:strings.Index(hotSlice, ">")], `checked=""`) {
		t.Fatalf("header = %q, want hot reply-sort-order checked by default", header)
	}
}

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
		`id="engagement-stats-root"`,
		`class="post-engagement-stats"`,
		`open=""`,
		`id="like-stats-root"`,
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

func TestPostPage_RendersFooterTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 25, 14, 56, 0, 0, time.UTC)
	createdAt := now.Add(-2 * time.Hour)

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:           "root",
			AuthorHandle: "bsky.app",
			Text:         "hello",
			CreatedAt:    createdAt,
		},
	}, now, nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="post-page-timestamp"`) {
		t.Fatalf("html = %q, want footer timestamp", html)
	}
	if !strings.Contains(html, "2026-07-25 12:56 UTC (2 hours ago)") {
		t.Fatalf("html = %q, want expanded footer timestamp", html)
	}

	pageIdx := strings.Index(html, `class="post post-page"`)
	if pageIdx < 0 {
		t.Fatalf("html = %q, want post-page article", html)
	}
	pageHTML := html[pageIdx:]
	headerStart := strings.Index(pageHTML, "<header>")
	headerEnd := strings.Index(pageHTML, "</header>")
	if headerStart < 0 || headerEnd < 0 {
		t.Fatalf("html = %q, want post-page header", html)
	}
	header := pageHTML[headerStart:headerEnd]
	if strings.Contains(header, "<time") {
		t.Fatalf("header = %q, want no byline timestamp on post page", header)
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
	repliesIdx := strings.Index(html, `id="post-replies-root"`)
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

func TestPostPage_RendersEmptyRepliesContainer(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:           "root",
			AuthorHandle: "bsky.app",
			Text:         "lonely post",
		},
	}, time.Now().UTC(), nil, "").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `id="post-replies-root"`) {
		t.Fatalf("html = %q, want empty replies container for live OOB swap target", html)
	}
}

func TestRepliesRefreshFragment_EmptyWhenAllKnown(t *testing.T) {
	t.Parallel()

	view := feedquery.PostPageView{
		Post: feedquery.PostView{ID: "root", AuthorHandle: "bsky.app"},
		Replies: []feedquery.ThreadNodeView{
			{Post: feedquery.PostView{ID: "reply1", AuthorHandle: "dev.example", Text: "reply one"}},
		},
	}

	var buf bytes.Buffer
	if err := post.RepliesRefreshFragment(view, map[string]bool{"reply1": true}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}
	if got := buf.String(); got != "" {
		t.Fatalf("html = %q, want empty when all replies are known", got)
	}
}

func TestRepliesRefreshFragment_FullListOOBWhenUnknown(t *testing.T) {
	t.Parallel()

	view := feedquery.PostPageView{
		Post: feedquery.PostView{ID: "root", AuthorHandle: "bsky.app"},
		Replies: []feedquery.ThreadNodeView{
			{Post: feedquery.PostView{ID: "reply1", AuthorHandle: "dev.example", Text: "reply one"}},
			{Post: feedquery.PostView{ID: "reply2", AuthorHandle: "dev.example", Text: "reply two"}},
		},
	}

	var buf bytes.Buffer
	if err := post.RepliesRefreshFragment(view, map[string]bool{"reply1": true}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="post-replies-root"`,
		`hx-swap-oob="true"`,
		`id="post-reply1"`,
		`id="post-reply2"`,
		"reply one",
		"reply two",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestRepliesRefreshFragment_FirstReplyIntoEmptyKnown(t *testing.T) {
	t.Parallel()

	view := feedquery.PostPageView{
		Post: feedquery.PostView{ID: "root", AuthorHandle: "bsky.app"},
		Replies: []feedquery.ThreadNodeView{
			{Post: feedquery.PostView{ID: "reply1", AuthorHandle: "dev.example", Text: "first reply"}},
		},
	}

	var buf bytes.Buffer
	if err := post.RepliesRefreshFragment(view, map[string]bool{}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `id="post-replies-root"`) || !strings.Contains(html, "first reply") {
		t.Fatalf("html = %q, want first reply into previously empty container", html)
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

func TestPostPage_RendersFilterTextForFilteredPost(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "abc",
			AuthorHandle:      "bsky.app",
			AuthorDisplayName: "Bluesky",
			Text:              "hidden content",
			Moderation: feedquery.ModerationView{
				Filtered:   true,
				FilterText: "Adult content",
			},
		},
	}, time.Now().UTC(), nil, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		"Adult content",
		`property="og:description" content="Adult content on Twisky"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, "Post hidden by moderation") {
		t.Fatalf("html = %q, want specific filter text instead of fallback", html)
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
