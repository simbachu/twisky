package post_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
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
	}, time.Now().UTC()).Render(&buf); err != nil {
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

func TestPostPage_OmitsAncestorsWithoutThreadContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := post.PostPage(feedquery.PostPageView{
		Post: feedquery.PostView{
			ID:                "root",
			AuthorHandle:      "bsky.app",
			AuthorDisplayName: "Bluesky",
		},
	}, time.Now().UTC()).Render(&buf); err != nil {
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
	}, time.Now().UTC()).Render(&buf); err != nil {
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
