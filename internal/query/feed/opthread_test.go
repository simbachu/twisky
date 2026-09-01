package feed_test

import (
	"encoding/json"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestNewPostViewFromFeedItem_OPThreadNumbering(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostViewFromFeedItem(bluesky.FeedItem{
		OPThreadPostIndex: 3,
		OPThreadPostCount: 6,
		Post: bluesky.Post{
			URI:    "at://did:plc:alice/app.bsky.feed.post/p3",
			Author: bluesky.Author{Handle: "alice.test", DID: "did:plc:alice"},
			Record: bluesky.PostRecord{
				Text: "third",
				Reply: &bluesky.RecordReplyRef{
					Root:   bluesky.StrongRef{URI: "at://did:plc:alice/app.bsky.feed.post/p1"},
					Parent: bluesky.StrongRef{URI: "at://did:plc:alice/app.bsky.feed.post/p2"},
				},
			},
		},
		Reply: &bluesky.ReplyContext{
			Root: &bluesky.Post{
				URI:    "at://did:plc:alice/app.bsky.feed.post/p1",
				Author: bluesky.Author{Handle: "alice.test", DID: "did:plc:alice"},
			},
		},
	})

	index, count, ok := view.OPThreadNumber()
	if !ok {
		t.Fatal("OPThreadNumber() ok = false, want true")
	}
	if index != 3 || count != 6 {
		t.Fatalf("OPThreadNumber() = %d/%d, want 3/6", index, count)
	}
	handle, did, rootID := view.ThreadRootPathParts()
	if rootID != "p1" || did != "did:plc:alice" || handle != "alice.test" {
		t.Fatalf("ThreadRootPathParts() = %q %q %q, want alice.test did:plc:alice p1", handle, did, rootID)
	}
}

func TestNewPostViewFromFeedItem_InvalidOPThreadNumberingOmitted(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostViewFromFeedItem(bluesky.FeedItem{
		OPThreadPostIndex: 7,
		OPThreadPostCount: 6,
		Post: bluesky.Post{
			URI:    "at://did:plc:alice/app.bsky.feed.post/p7",
			Author: bluesky.Author{Handle: "alice.test", DID: "did:plc:alice"},
		},
	})

	if _, _, ok := view.OPThreadNumber(); ok {
		t.Fatal("OPThreadNumber() ok = true, want false for index > count")
	}
}

func TestFeedItem_UnmarshalJSON_OPThreadFields(t *testing.T) {
	t.Parallel()

	const raw = `{
		"post": {
			"uri": "at://did:plc:alice/app.bsky.feed.post/p2",
			"author": {"did": "did:plc:alice", "handle": "alice.test"},
			"record": {"text": "second", "createdAt": "2026-01-15T12:00:00.000Z"}
		},
		"opThreadPostIndex": 2,
		"opThreadPostCount": 5,
		"reply": {
			"root": {
				"uri": "at://did:plc:alice/app.bsky.feed.post/p1",
				"author": {"did": "did:plc:alice", "handle": "alice.test"},
				"record": {"text": "first", "createdAt": "2026-01-15T11:00:00.000Z"}
			},
			"parent": {
				"uri": "at://did:plc:alice/app.bsky.feed.post/p1",
				"author": {"did": "did:plc:alice", "handle": "alice.test"},
				"record": {"text": "first", "createdAt": "2026-01-15T11:00:00.000Z"}
			}
		}
	}`

	var item bluesky.FeedItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatalf("Unmarshal() err = %v", err)
	}
	if item.OPThreadPostIndex != 2 || item.OPThreadPostCount != 5 {
		t.Fatalf("numbering = %d/%d, want 2/5", item.OPThreadPostIndex, item.OPThreadPostCount)
	}
	if item.Reply == nil || item.Reply.Root == nil || item.Reply.Root.URI == "" {
		t.Fatal("reply.root not parsed")
	}
}
