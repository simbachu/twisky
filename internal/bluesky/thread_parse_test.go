package bluesky

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseThreadNode_NotFound(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{"$type":"app.bsky.feed.defs#notFoundPost","uri":"at://did:plc:x/app.bsky.feed.post/abc"}`)
	node, err := parseThreadNode(raw)
	if err != nil {
		t.Fatalf("parseThreadNode: %v", err)
	}
	nf, ok := node.(NotFoundPost)
	if !ok {
		t.Fatalf("type = %T, want NotFoundPost", node)
	}
	if nf.URI != "at://did:plc:x/app.bsky.feed.post/abc" {
		t.Fatalf("URI = %q", nf.URI)
	}
}

func TestParseThreadNode_Blocked(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{"$type":"app.bsky.feed.defs#blockedPost","uri":"at://did:plc:x/app.bsky.feed.post/abc"}`)
	node, err := parseThreadNode(raw)
	if err != nil {
		t.Fatalf("parseThreadNode: %v", err)
	}
	bp, ok := node.(BlockedPost)
	if !ok {
		t.Fatalf("type = %T, want BlockedPost", node)
	}
	if bp.URI != "at://did:plc:x/app.bsky.feed.post/abc" {
		t.Fatalf("URI = %q", bp.URI)
	}
}

func TestParseThreadNode_UnknownType(t *testing.T) {
	t.Parallel()

	_, err := parseThreadNode(json.RawMessage(`{"$type":"app.bsky.feed.defs#mystery"}`))
	if err == nil || !strings.Contains(err.Error(), "unknown thread node type") {
		t.Fatalf("err = %v, want unknown type", err)
	}
}

func TestParseThreadNode_EmptyRaw(t *testing.T) {
	t.Parallel()

	node, err := parseThreadNode(nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if node != nil {
		t.Fatalf("node = %#v, want nil", node)
	}
}

func TestParseThreadNode_NestedParentAndReplies(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{
		"$type":"app.bsky.feed.defs#threadViewPost",
		"post":{"uri":"at://did:plc:child/app.bsky.feed.post/c1","cid":"cidc","author":{"did":"did:plc:child","handle":"child.test"},"record":{"text":"child","createdAt":"2026-01-01T00:00:00Z"}},
		"parent":{"$type":"app.bsky.feed.defs#notFoundPost","uri":"at://did:plc:parent/app.bsky.feed.post/p1"},
		"replies":[
			{"$type":"app.bsky.feed.defs#blockedPost","uri":"at://did:plc:reply/app.bsky.feed.post/r1"},
			{"$type":"app.bsky.feed.defs#threadViewPost","post":{"uri":"at://did:plc:reply/app.bsky.feed.post/r2","cid":"cidr","author":{"did":"did:plc:reply","handle":"reply.test"},"record":{"text":"ok","createdAt":"2026-01-01T00:00:00Z"}}}
		]
	}`)
	node, err := parseThreadNode(raw)
	if err != nil {
		t.Fatalf("parseThreadNode: %v", err)
	}
	tv, ok := node.(ThreadViewPost)
	if !ok {
		t.Fatalf("type = %T", node)
	}
	if _, ok := tv.Parent.(NotFoundPost); !ok {
		t.Fatalf("Parent = %T, want NotFoundPost", tv.Parent)
	}
	if len(tv.Replies) != 2 {
		t.Fatalf("Replies = %d, want 2", len(tv.Replies))
	}
	if _, ok := tv.Replies[0].(BlockedPost); !ok {
		t.Fatalf("Replies[0] = %T, want BlockedPost", tv.Replies[0])
	}
	reply, ok := tv.Replies[1].(ThreadViewPost)
	if !ok {
		t.Fatalf("Replies[1] = %T, want ThreadViewPost", tv.Replies[1])
	}
	if reply.Post.Record.Text != "ok" {
		t.Fatalf("reply text = %q", reply.Post.Record.Text)
	}
}
