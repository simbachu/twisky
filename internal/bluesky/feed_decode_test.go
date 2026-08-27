package bluesky_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
)

const mixedTimelineFeedJSON = `{
	"feed": [
		{
			"post": {
				"uri": "at://did:plc:good/app.bsky.feed.post/good1",
				"author": {"did": "did:plc:good", "handle": "good.test"},
				"record": {
					"text": "valid post",
					"createdAt": "2026-01-15T12:00:00.000Z"
				}
			}
		},
		{
			"post": {
				"uri": "at://did:plc:badtime/app.bsky.feed.post/bad1",
				"author": {"did": "did:plc:badtime", "handle": "badtime.test"},
				"record": {
					"text": "empty createdAt",
					"createdAt": ""
				}
			}
		},
		{
			"post": {
				"uri": "at://did:plc:nulltime/app.bsky.feed.post/null1",
				"author": {"did": "did:plc:nulltime", "handle": "nulltime.test"},
				"record": {
					"text": "null createdAt",
					"createdAt": null
				}
			}
		},
		{
			"post": {
				"uri": "at://did:plc:pinned/app.bsky.feed.post/pin1",
				"author": {"did": "did:plc:pinned", "handle": "pinned.test"},
				"record": {
					"text": "pinned post",
					"createdAt": "2026-01-14T12:00:00.000Z"
				}
			},
			"reason": {"$type": "app.bsky.feed.defs#reasonPin"}
		},
		{
			"post": {
				"uri": "at://did:plc:reply/app.bsky.feed.post/reply1",
				"author": {"did": "did:plc:reply", "handle": "reply.test"},
				"record": {
					"text": "reply with missing parent",
					"createdAt": "2026-01-13T12:00:00.000Z",
					"reply": {
						"root": {"uri": "at://did:plc:root/app.bsky.feed.post/root1", "cid": "bafyroot"},
						"parent": {"uri": "at://did:plc:missing/app.bsky.feed.post/missing1", "cid": "bafyparent"}
					}
				}
			},
			"reply": {
				"root": {"uri": "at://did:plc:root/app.bsky.feed.post/root1", "cid": "bafyroot"},
				"parent": {
					"$type": "app.bsky.feed.defs#notFoundPost",
					"uri": "at://did:plc:missing/app.bsky.feed.post/missing1",
					"notFound": true
				}
			}
		},
		{
			"post": {
				"uri": "at://did:plc:broken/app.bsky.feed.post/broken1",
				"author": {"did": "did:plc:broken", "handle": "broken.test"},
				"record": {"text": "not an object"}
			},
			"reason": 42
		}
	],
	"cursor": "next-page"
}`

func TestParseFeedResponse_MixedTimelinePayload(t *testing.T) {
	t.Parallel()

	feed, cursor, err := bluesky.ParseFeedResponse([]byte(mixedTimelineFeedJSON))
	if err != nil {
		t.Fatalf("ParseFeedResponse() err = %v", err)
	}
	if cursor != "next-page" {
		t.Fatalf("cursor = %q, want next-page", cursor)
	}
	if len(feed) != 5 {
		t.Fatalf("len(feed) = %d, want 5 (malformed item skipped)", len(feed))
	}

	want := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	if !feed[0].Post.Record.CreatedAt.Equal(want) {
		t.Fatalf("feed[0].CreatedAt = %v, want %v", feed[0].Post.Record.CreatedAt, want)
	}
	if feed[0].Post.Record.Text != "valid post" {
		t.Fatalf("feed[0].Text = %q, want valid post", feed[0].Post.Record.Text)
	}

	if !feed[1].Post.Record.CreatedAt.IsZero() {
		t.Fatalf("feed[1].CreatedAt = %v, want zero for empty createdAt", feed[1].Post.Record.CreatedAt)
	}
	if feed[1].Post.Record.Text != "empty createdAt" {
		t.Fatalf("feed[1].Text = %q, want empty createdAt", feed[1].Post.Record.Text)
	}

	if !feed[2].Post.Record.CreatedAt.IsZero() {
		t.Fatalf("feed[2].CreatedAt = %v, want zero for null createdAt", feed[2].Post.Record.CreatedAt)
	}

	if feed[3].Reason != nil {
		t.Fatalf("feed[3].Reason = %#v, want nil for reasonPin", feed[3].Reason)
	}
	if feed[3].Post.Record.Text != "pinned post" {
		t.Fatalf("feed[3].Text = %q, want pinned post", feed[3].Post.Record.Text)
	}

	if feed[4].Reply == nil || feed[4].Reply.Parent != nil {
		t.Fatalf("feed[4].Reply = %#v, want reply with nil parent for notFoundPost", feed[4].Reply)
	}
}

func TestParseFeedResponse_AllOrNothingDecodeFailsOnBadCreatedAt(t *testing.T) {
	t.Parallel()

	var payload struct {
		Feed []bluesky.FeedItem `json:"feed"`
	}
	err := json.Unmarshal([]byte(mixedTimelineFeedJSON), &payload)
	if err == nil {
		t.Fatal("json.Unmarshal into []FeedItem err = nil, want error for empty createdAt")
	}
}
