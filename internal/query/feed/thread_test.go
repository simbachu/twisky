package feed_test

import (
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func threadWithAncestorsAndReplies() bluesky.ThreadViewPost {
	return bluesky.ThreadViewPost{
		Post: bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/root",
			Author: bluesky.Author{Handle: "bsky.app"},
			Record: bluesky.PostRecord{Text: "root"},
		},
		Parent: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/grandparent",
				Author: bluesky.Author{Handle: "bsky.app"},
				Record: bluesky.PostRecord{Text: "grandparent"},
			},
			Parent: bluesky.ThreadViewPost{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/parent",
					Author: bluesky.Author{Handle: "bsky.app"},
					Record: bluesky.PostRecord{Text: "parent"},
				},
			},
		},
		Replies: []bluesky.ThreadNode{
			bluesky.ThreadViewPost{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/reply",
					Author: bluesky.Author{Handle: "dev.example"},
					Record: bluesky.PostRecord{Text: "reply"},
				},
			},
			bluesky.NotFoundPost{URI: "at://did:plc:example/app.bsky.feed.post/missing"},
		},
	}
}

func TestNewPostPageView_FullPageOmitsAncestors(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostPageView(threadWithAncestorsAndReplies(), "")

	if view.Post.ID != "root" {
		t.Fatalf("view.Post.ID = %q, want root", view.Post.ID)
	}
	if !view.HasAncestors {
		t.Fatal("HasAncestors = false, want true")
	}
	if len(view.Ancestors) != 0 {
		t.Fatalf("len(view.Ancestors) = %d, want 0 on full page", len(view.Ancestors))
	}
	if len(view.Replies) != 2 {
		t.Fatalf("len(view.Replies) = %d, want 2", len(view.Replies))
	}
	if view.Replies[0].Post.ID != "reply" {
		t.Fatalf("view.Replies[0].Post.ID = %q, want reply", view.Replies[0].Post.ID)
	}
	if !view.Replies[1].Unavailable {
		t.Fatalf("view.Replies[1].Unavailable = false, want true")
	}
}

func TestNewPostPageView_AncestorsFragment(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostPageView(threadWithAncestorsAndReplies(), feedquery.PostPagePartAncestors)

	if view.Post.ID != "" {
		t.Fatalf("view.Post.ID = %q, want empty on ancestors fragment", view.Post.ID)
	}
	if len(view.Ancestors) != 2 {
		t.Fatalf("len(view.Ancestors) = %d, want 2", len(view.Ancestors))
	}
	if view.Ancestors[0].Post.ID != "parent" || view.Ancestors[1].Post.ID != "grandparent" {
		t.Fatalf("view.Ancestors = %#v, want parent then grandparent", view.Ancestors)
	}
	if len(view.Replies) != 0 {
		t.Fatalf("len(view.Replies) = %d, want 0 on ancestors fragment", len(view.Replies))
	}
}

func TestNewThreadNodeViews_SortsRepliesOldestFirst(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	view := feedquery.NewPostPageView(bluesky.ThreadViewPost{
		Post: bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/root",
			Author: bluesky.Author{Handle: "bsky.app"},
			Record: bluesky.PostRecord{Text: "root"},
		},
		Replies: []bluesky.ThreadNode{
			bluesky.ThreadViewPost{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/newer",
					Author: bluesky.Author{Handle: "dev.example"},
					Record: bluesky.PostRecord{Text: "newer", CreatedAt: newer},
				},
			},
			bluesky.ThreadViewPost{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/older",
					Author: bluesky.Author{Handle: "dev.example"},
					Record: bluesky.PostRecord{Text: "older", CreatedAt: older},
				},
				Replies: []bluesky.ThreadNode{
					bluesky.ThreadViewPost{
						Post: bluesky.Post{
							URI:    "at://did:plc:example/app.bsky.feed.post/nested-newer",
							Author: bluesky.Author{Handle: "dev.example"},
							Record: bluesky.PostRecord{Text: "nested newer", CreatedAt: newer},
						},
					},
					bluesky.ThreadViewPost{
						Post: bluesky.Post{
							URI:    "at://did:plc:example/app.bsky.feed.post/nested-older",
							Author: bluesky.Author{Handle: "dev.example"},
							Record: bluesky.PostRecord{Text: "nested older", CreatedAt: older},
						},
					},
				},
			},
		},
	}, "")

	if len(view.Replies) != 2 {
		t.Fatalf("len(view.Replies) = %d, want 2", len(view.Replies))
	}
	if view.Replies[0].Post.ID != "older" {
		t.Fatalf("view.Replies[0].Post.ID = %q, want older", view.Replies[0].Post.ID)
	}
	if view.Replies[1].Post.ID != "newer" {
		t.Fatalf("view.Replies[1].Post.ID = %q, want newer", view.Replies[1].Post.ID)
	}
	if len(view.Replies[0].Replies) != 2 {
		t.Fatalf("len(view.Replies[0].Replies) = %d, want 2", len(view.Replies[0].Replies))
	}
	if view.Replies[0].Replies[0].Post.ID != "nested-older" {
		t.Fatalf("view.Replies[0].Replies[0].Post.ID = %q, want nested-older", view.Replies[0].Replies[0].Post.ID)
	}
	if view.Replies[0].Replies[1].Post.ID != "nested-newer" {
		t.Fatalf("view.Replies[0].Replies[1].Post.ID = %q, want nested-newer", view.Replies[0].Replies[1].Post.ID)
	}
}

func TestNewPostPageView_BlockedParent(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostPageView(bluesky.ThreadViewPost{
		Post: bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/reply",
			Author: bluesky.Author{Handle: "dev.example"},
			Record: bluesky.PostRecord{Text: "reply"},
		},
		Parent: bluesky.BlockedPost{URI: "at://did:plc:example/app.bsky.feed.post/parent"},
	}, "")

	if !view.HasAncestors {
		t.Fatal("HasAncestors = false, want true for blocked parent")
	}
	if len(view.Ancestors) != 0 {
		t.Fatalf("len(view.Ancestors) = %d, want 0 on full page", len(view.Ancestors))
	}

	fragment := feedquery.NewPostPageView(bluesky.ThreadViewPost{
		Post: bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/reply",
			Author: bluesky.Author{Handle: "dev.example"},
			Record: bluesky.PostRecord{Text: "reply"},
		},
		Parent: bluesky.BlockedPost{URI: "at://did:plc:example/app.bsky.feed.post/parent"},
	}, feedquery.PostPagePartAncestors)

	if len(fragment.Ancestors) != 1 {
		t.Fatalf("len(fragment.Ancestors) = %d, want 1", len(fragment.Ancestors))
	}
	if !fragment.Ancestors[0].Unavailable {
		t.Fatal("fragment.Ancestors[0].Unavailable = false, want true")
	}
}

func TestNewPostPageView_NotFoundParent(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostPageView(bluesky.ThreadViewPost{
		Post: bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/reply",
			Author: bluesky.Author{Handle: "dev.example"},
			Record: bluesky.PostRecord{Text: "reply"},
		},
		Parent: bluesky.NotFoundPost{URI: "at://did:plc:example/app.bsky.feed.post/parent"},
	}, feedquery.PostPagePartAncestors)

	if len(view.Ancestors) != 1 || !view.Ancestors[0].Unavailable {
		t.Fatalf("view.Ancestors = %#v, want one unavailable ancestor", view.Ancestors)
	}
}
