package repost_test

import (
	"context"
	"errors"
	"testing"

	"github.com/simbachu/twisky/internal/command/repost"
	"github.com/simbachu/twisky/internal/intent"
)

type stubWriter struct {
	uri, cid, record string
	recordURI        string
	err              error
	calls            int
	deleteCalls      int
}

func (s *stubWriter) CreateRepost(_ context.Context, uri, cid string) (string, error) {
	s.calls++
	s.uri = uri
	s.cid = cid
	return s.recordURI, s.err
}

func (s *stubWriter) DeleteRepost(_ context.Context, recordURI string) error {
	s.deleteCalls++
	s.record = recordURI
	return s.err
}

func TestHandler_CreateRepost(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{recordURI: "at://did:plc:me/app.bsky.feed.repost/r1"}
	h := repost.NewHandler(writer)

	got, err := h.HandleRepost(context.Background(), intent.RepostPost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	})
	if err != nil {
		t.Fatalf("HandleRepost() err = %v", err)
	}
	if writer.calls != 1 {
		t.Fatalf("calls = %d, want 1", writer.calls)
	}
	if writer.uri != "at://did:plc:example/app.bsky.feed.post/abc" || writer.cid != "bafycid" {
		t.Fatalf("CreateRepost(%q, %q)", writer.uri, writer.cid)
	}
	if got != writer.recordURI {
		t.Fatalf("record URI = %q, want %q", got, writer.recordURI)
	}
}

func TestHandler_Unrepost(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	h := repost.NewHandler(writer)

	err := h.HandleUnrepost(context.Background(), intent.UnrepostPost{
		Record: "at://did:plc:me/app.bsky.feed.repost/r1",
	})
	if err != nil {
		t.Fatalf("HandleUnrepost() err = %v", err)
	}
	if writer.deleteCalls != 1 || writer.record != "at://did:plc:me/app.bsky.feed.repost/r1" {
		t.Fatalf("DeleteRepost calls=%d record=%q", writer.deleteCalls, writer.record)
	}
}

func TestHandler_RejectsEmptyURIOrCID(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	h := repost.NewHandler(writer)

	if _, err := h.HandleRepost(context.Background(), intent.RepostPost{URI: "", CID: "bafycid"}); err == nil {
		t.Fatal("HandleRepost() empty URI err = nil, want error")
	}
	if _, err := h.HandleRepost(context.Background(), intent.RepostPost{URI: "at://x", CID: ""}); err == nil {
		t.Fatal("HandleRepost() empty CID err = nil, want error")
	}
	if writer.calls != 0 {
		t.Fatalf("calls = %d, want 0", writer.calls)
	}
}

func TestHandler_UnrepostRejectsEmptyRecord(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	h := repost.NewHandler(writer)

	if err := h.HandleUnrepost(context.Background(), intent.UnrepostPost{Record: ""}); err == nil {
		t.Fatal("HandleUnrepost() empty record err = nil, want error")
	}
	if writer.deleteCalls != 0 {
		t.Fatalf("deleteCalls = %d, want 0", writer.deleteCalls)
	}
}

func TestHandler_PropagatesWriterError(t *testing.T) {
	t.Parallel()

	want := errors.New("pds down")
	writer := &stubWriter{err: want}
	h := repost.NewHandler(writer)

	_, err := h.HandleRepost(context.Background(), intent.RepostPost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	})
	if !errors.Is(err, want) {
		t.Fatalf("HandleRepost() err = %v, want %v", err, want)
	}
}
