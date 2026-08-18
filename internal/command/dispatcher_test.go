package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/simbachu/twisky/internal/command"
	"github.com/simbachu/twisky/internal/intent"
)

type stubLikeWriter struct {
	calls int
	err   error
}

func (s *stubLikeWriter) CreateLike(context.Context, string, string) error {
	s.calls++
	return s.err
}

type stubRepostWriter struct {
	calls int
	err   error
	uri   string
	cid   string
}

func (s *stubRepostWriter) CreateRepost(_ context.Context, uri, cid string) error {
	s.calls++
	s.uri = uri
	s.cid = cid
	return s.err
}

func TestDispatcher_LikePost(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{}
	d := command.NewDispatcher()
	err := d.Dispatch(context.Background(), intent.LikePost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	}, writer)
	if err != nil {
		t.Fatalf("Dispatch() err = %v", err)
	}
	if writer.calls != 1 {
		t.Fatalf("calls = %d, want 1", writer.calls)
	}
}

func TestDispatcher_RepostPost(t *testing.T) {
	t.Parallel()

	writer := &stubRepostWriter{}
	d := command.NewDispatcher()
	err := d.Dispatch(context.Background(), intent.RepostPost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	}, writer)
	if err != nil {
		t.Fatalf("Dispatch() err = %v", err)
	}
	if writer.calls != 1 || writer.uri == "" || writer.cid != "bafycid" {
		t.Fatalf("CreateRepost calls=%d uri=%q cid=%q", writer.calls, writer.uri, writer.cid)
	}
}

func TestDispatcher_UnknownIntent(t *testing.T) {
	t.Parallel()

	d := command.NewDispatcher()
	err := d.Dispatch(context.Background(), intent.ViewProfile{Slug: "x"}, &stubLikeWriter{})
	if err == nil {
		t.Fatal("Dispatch() err = nil, want unknown intent error")
	}
}

func TestDispatcher_PropagatesWriterError(t *testing.T) {
	t.Parallel()

	want := errors.New("boom")
	d := command.NewDispatcher()
	err := d.Dispatch(context.Background(), intent.LikePost{
		URI: "at://did:plc:example/app.bsky.feed.post/abc",
		CID: "bafycid",
	}, &stubLikeWriter{err: want})
	if !errors.Is(err, want) {
		t.Fatalf("Dispatch() err = %v, want %v", err, want)
	}
}
