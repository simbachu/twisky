package like_test

import (
	"context"
	"errors"
	"testing"

	"github.com/simbachu/twisky/internal/command/like"
	"github.com/simbachu/twisky/internal/intent"
)

type stubWriter struct {
	uri, cid string
	err      error
	calls    int
}

func (s *stubWriter) CreateLike(_ context.Context, uri, cid string) error {
	s.calls++
	s.uri = uri
	s.cid = cid
	return s.err
}

func TestHandler_CreateLike(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	h := like.NewHandler(writer)

	err := h.Handle(context.Background(), intent.LikePost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	})
	if err != nil {
		t.Fatalf("Handle() err = %v", err)
	}
	if writer.calls != 1 {
		t.Fatalf("calls = %d, want 1", writer.calls)
	}
	if writer.uri != "at://did:plc:example/app.bsky.feed.post/abc" || writer.cid != "bafycid" {
		t.Fatalf("CreateLike(%q, %q)", writer.uri, writer.cid)
	}
}

func TestHandler_RejectsEmptyURIOrCID(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	h := like.NewHandler(writer)

	if err := h.Handle(context.Background(), intent.LikePost{URI: "", CID: "bafycid"}); err == nil {
		t.Fatal("Handle() empty URI err = nil, want error")
	}
	if err := h.Handle(context.Background(), intent.LikePost{URI: "at://x", CID: ""}); err == nil {
		t.Fatal("Handle() empty CID err = nil, want error")
	}
	if writer.calls != 0 {
		t.Fatalf("calls = %d, want 0", writer.calls)
	}
}

func TestHandler_PropagatesWriterError(t *testing.T) {
	t.Parallel()

	want := errors.New("pds down")
	writer := &stubWriter{err: want}
	h := like.NewHandler(writer)

	err := h.Handle(context.Background(), intent.LikePost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	})
	if !errors.Is(err, want) {
		t.Fatalf("Handle() err = %v, want %v", err, want)
	}
}
