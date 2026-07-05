package feed_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestNewPostPageView_CollectsAncestorsAndReplies(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostPageView(bluesky.ThreadViewPost{
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
	})

	if view.Post.ID != "root" {
		t.Fatalf("view.Post.ID = %q, want root", view.Post.ID)
	}
	if len(view.Ancestors) != 2 {
		t.Fatalf("len(view.Ancestors) = %d, want 2", len(view.Ancestors))
	}
	if view.Ancestors[0].ID != "parent" || view.Ancestors[1].ID != "grandparent" {
		t.Fatalf("view.Ancestors = %#v, want parent then grandparent", view.Ancestors)
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
