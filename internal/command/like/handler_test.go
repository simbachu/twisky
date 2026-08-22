package like_test

import (
	"context"
	"errors"
	"testing"

	"github.com/simbachu/twisky/internal/command/like"
	"github.com/simbachu/twisky/internal/intent"
)

type stubWriter struct {
	uri, cid, record string
	recordURI        string
	err              error
	calls            int
	deleteCalls      int
}

func (s *stubWriter) CreateLike(_ context.Context, uri, cid string) (string, error) {
	s.calls++
	s.uri = uri
	s.cid = cid
	return s.recordURI, s.err
}

func (s *stubWriter) DeleteLike(_ context.Context, recordURI string) error {
	s.deleteCalls++
	s.record = recordURI
	return s.err
}

func TestHandler_CreateLike(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{recordURI: "at://did:plc:me/app.bsky.feed.like/like1"}
	h := like.NewHandler(writer)

	got, err := h.HandleLike(context.Background(), intent.LikePost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	})
	if err != nil {
		t.Fatalf("HandleLike() err = %v", err)
	}
	if writer.calls != 1 {
		t.Fatalf("calls = %d, want 1", writer.calls)
	}
	if writer.uri != "at://did:plc:example/app.bsky.feed.post/abc" || writer.cid != "bafycid" {
		t.Fatalf("CreateLike(%q, %q)", writer.uri, writer.cid)
	}
	if got != writer.recordURI {
		t.Fatalf("record URI = %q, want %q", got, writer.recordURI)
	}
}

func TestHandler_Unlike(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	h := like.NewHandler(writer)

	err := h.HandleUnlike(context.Background(), intent.UnlikePost{
		Record: "at://did:plc:me/app.bsky.feed.like/like1",
	})
	if err != nil {
		t.Fatalf("HandleUnlike() err = %v", err)
	}
	if writer.deleteCalls != 1 || writer.record != "at://did:plc:me/app.bsky.feed.like/like1" {
		t.Fatalf("DeleteLike calls=%d record=%q", writer.deleteCalls, writer.record)
	}
}

func TestHandler_RejectsEmptyURIOrCID(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	h := like.NewHandler(writer)

	if _, err := h.HandleLike(context.Background(), intent.LikePost{URI: "", CID: "bafycid"}); err == nil {
		t.Fatal("HandleLike() empty URI err = nil, want error")
	}
	if _, err := h.HandleLike(context.Background(), intent.LikePost{URI: "at://x", CID: ""}); err == nil {
		t.Fatal("HandleLike() empty CID err = nil, want error")
	}
	if writer.calls != 0 {
		t.Fatalf("calls = %d, want 0", writer.calls)
	}
}

func TestHandler_UnlikeRejectsEmptyRecord(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	h := like.NewHandler(writer)

	if err := h.HandleUnlike(context.Background(), intent.UnlikePost{Record: ""}); err == nil {
		t.Fatal("HandleUnlike() empty record err = nil, want error")
	}
	if writer.deleteCalls != 0 {
		t.Fatalf("deleteCalls = %d, want 0", writer.deleteCalls)
	}
}

func TestHandler_PropagatesWriterError(t *testing.T) {
	t.Parallel()

	want := errors.New("pds down")
	writer := &stubWriter{err: want}
	h := like.NewHandler(writer)

	_, err := h.HandleLike(context.Background(), intent.LikePost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	})
	if !errors.Is(err, want) {
		t.Fatalf("HandleLike() err = %v, want %v", err, want)
	}
}
