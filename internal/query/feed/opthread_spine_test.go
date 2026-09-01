package feed_test

import (
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestCollectOPThoughtSpine(t *testing.T) {
	t.Parallel()

	root := bluesky.ThreadViewPost{
		Post: bluesky.Post{
			URI:    "at://did:plc:alice/app.bsky.feed.post/p1",
			Author: bluesky.Author{DID: "did:plc:alice", Handle: "alice.test"},
			Record: bluesky.PostRecord{
				Text:      "one",
				CreatedAt: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			},
		},
		Replies: []bluesky.ThreadNode{
			bluesky.ThreadViewPost{
				Post: bluesky.Post{
					URI:    "at://did:plc:bob/app.bsky.feed.post/r1",
					Author: bluesky.Author{DID: "did:plc:bob", Handle: "bob.test"},
					Record: bluesky.PostRecord{Text: "bob"},
				},
			},
			bluesky.ThreadViewPost{
				Post: bluesky.Post{
					URI:    "at://did:plc:alice/app.bsky.feed.post/p2",
					Author: bluesky.Author{DID: "did:plc:alice", Handle: "alice.test"},
					Record: bluesky.PostRecord{
						Text:      "two",
						CreatedAt: time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC),
					},
				},
				Replies: []bluesky.ThreadNode{
					bluesky.ThreadViewPost{
						Post: bluesky.Post{
							URI:    "at://did:plc:alice/app.bsky.feed.post/p3",
							Author: bluesky.Author{DID: "did:plc:alice", Handle: "alice.test"},
							Record: bluesky.PostRecord{
								Text:      "three",
								CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
							},
						},
					},
				},
			},
		},
	}

	spine := feedquery.CollectOPThoughtSpine(root)
	if len(spine) != 3 {
		t.Fatalf("len(spine) = %d, want 3", len(spine))
	}
	for i, want := range []string{"p1", "p2", "p3"} {
		if spine[i].ID != want {
			t.Fatalf("spine[%d].ID = %q, want %q", i, spine[i].ID, want)
		}
	}
}
