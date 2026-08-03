package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/simbachu/twisky/internal/command"
	"github.com/simbachu/twisky/internal/intent"
)

type stubWriter struct {
	calls int
	err   error
}

func (s *stubWriter) CreateLike(context.Context, string, string) error {
	s.calls++
	return s.err
}

func TestDispatcher_LikePost(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
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

func TestDispatcher_UnknownIntent(t *testing.T) {
	t.Parallel()

	d := command.NewDispatcher()
	err := d.Dispatch(context.Background(), intent.ViewProfile{Slug: "x"}, &stubWriter{})
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
	}, &stubWriter{err: want})
	if !errors.Is(err, want) {
		t.Fatalf("Dispatch() err = %v, want %v", err, want)
	}
}
