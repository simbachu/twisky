package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/simbachu/twisky/internal/command"
	"github.com/simbachu/twisky/internal/intent"
)

type stubLikeWriter struct {
	calls       int
	deleteCalls int
	err         error
	recordURI   string
}

func (s *stubLikeWriter) CreateLike(context.Context, string, string) (string, error) {
	s.calls++
	return s.recordURI, s.err
}

func (s *stubLikeWriter) DeleteLike(context.Context, string) error {
	s.deleteCalls++
	return s.err
}

type stubRepostWriter struct {
	calls       int
	deleteCalls int
	err         error
	uri         string
	cid         string
	recordURI   string
}

func (s *stubRepostWriter) CreateRepost(_ context.Context, uri, cid string) (string, error) {
	s.calls++
	s.uri = uri
	s.cid = cid
	return s.recordURI, s.err
}

func (s *stubRepostWriter) DeleteRepost(context.Context, string) error {
	s.deleteCalls++
	return s.err
}

func TestDispatcher_LikePost(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{recordURI: "at://did:plc:me/app.bsky.feed.like/like1"}
	d := command.NewDispatcher()
	got, err := d.Dispatch(context.Background(), intent.LikePost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	}, writer)
	if err != nil {
		t.Fatalf("Dispatch() err = %v", err)
	}
	if writer.calls != 1 {
		t.Fatalf("calls = %d, want 1", writer.calls)
	}
	if got != writer.recordURI {
		t.Fatalf("record URI = %q, want %q", got, writer.recordURI)
	}
}

func TestDispatcher_UnlikePost(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{}
	d := command.NewDispatcher()
	got, err := d.Dispatch(context.Background(), intent.UnlikePost{
		Record: "at://did:plc:me/app.bsky.feed.like/like1",
	}, writer)
	if err != nil {
		t.Fatalf("Dispatch() err = %v", err)
	}
	if got != "" {
		t.Fatalf("record URI = %q, want empty", got)
	}
	if writer.deleteCalls != 1 {
		t.Fatalf("deleteCalls = %d, want 1", writer.deleteCalls)
	}
}

func TestDispatcher_RepostPost(t *testing.T) {
	t.Parallel()

	writer := &stubRepostWriter{recordURI: "at://did:plc:me/app.bsky.feed.repost/r1"}
	d := command.NewDispatcher()
	got, err := d.Dispatch(context.Background(), intent.RepostPost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	}, writer)
	if err != nil {
		t.Fatalf("Dispatch() err = %v", err)
	}
	if writer.calls != 1 || writer.uri == "" || writer.cid != "bafycid" {
		t.Fatalf("CreateRepost calls=%d uri=%q cid=%q", writer.calls, writer.uri, writer.cid)
	}
	if got != writer.recordURI {
		t.Fatalf("record URI = %q, want %q", got, writer.recordURI)
	}
}

func TestDispatcher_UnrepostPost(t *testing.T) {
	t.Parallel()

	writer := &stubRepostWriter{}
	d := command.NewDispatcher()
	got, err := d.Dispatch(context.Background(), intent.UnrepostPost{
		Record: "at://did:plc:me/app.bsky.feed.repost/r1",
	}, writer)
	if err != nil {
		t.Fatalf("Dispatch() err = %v", err)
	}
	if got != "" {
		t.Fatalf("record URI = %q, want empty", got)
	}
	if writer.deleteCalls != 1 {
		t.Fatalf("deleteCalls = %d, want 1", writer.deleteCalls)
	}
}

func TestDispatcher_UnknownIntent(t *testing.T) {
	t.Parallel()

	d := command.NewDispatcher()
	_, err := d.Dispatch(context.Background(), intent.ViewProfile{Slug: "x"}, &stubLikeWriter{})
	if err == nil {
		t.Fatal("Dispatch() err = nil, want unknown intent error")
	}
}

func TestDispatcher_PropagatesWriterError(t *testing.T) {
	t.Parallel()

	want := errors.New("boom")
	d := command.NewDispatcher()
	_, err := d.Dispatch(context.Background(), intent.LikePost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	}, &stubLikeWriter{err: want})
	if !errors.Is(err, want) {
		t.Fatalf("Dispatch() err = %v, want %v", err, want)
	}
}
