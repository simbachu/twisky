package post_test

import (
	"context"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/command/post"
	"github.com/simbachu/twisky/internal/intent"
)

type stubWriter struct {
	text      string
	reply     *intent.ReplyTo
	recordURI string
	err       error
}

func (s *stubWriter) CreatePost(_ context.Context, text string, reply *intent.ReplyTo) (string, error) {
	s.text = text
	s.reply = reply
	if s.recordURI == "" {
		s.recordURI = "at://did:plc:me/app.bsky.feed.post/abc123"
	}
	return s.recordURI, s.err
}

func TestHandler_CreatePost_Success(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	uri, err := handler.HandleCreate(context.Background(), intent.CreatePost{Text: "  hello  "})
	if err != nil {
		t.Fatalf("HandleCreate() err = %v", err)
	}
	if uri != writer.recordURI {
		t.Fatalf("uri = %q, want %q", uri, writer.recordURI)
	}
	if writer.text != "hello" {
		t.Fatalf("text = %q, want trimmed hello", writer.text)
	}
	if writer.reply != nil {
		t.Fatalf("reply = %#v, want nil", writer.reply)
	}
}

func TestHandler_CreatePost_EmptyRejected(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{Text: "   "})
	if err == nil {
		t.Fatal("HandleCreate() err = nil, want validation error")
	}
	if writer.text != "" {
		t.Fatalf("writer called with %q, want no call", writer.text)
	}
}

func TestHandler_CreatePost_TooLongRejected(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	text := strings.Repeat("a", post.MaxPostGraphemes+1)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{Text: text})
	if err == nil {
		t.Fatal("HandleCreate() err = nil, want validation error")
	}
	if writer.text != "" {
		t.Fatalf("writer called with text, want no call")
	}
}

func TestHandler_CreatePost_ForwardsReply(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	reply := &intent.ReplyTo{
		RootURI:   "at://did:plc:root/app.bsky.feed.post/root1",
		RootCID:   "bafyroot",
		ParentURI: "at://did:plc:parent/app.bsky.feed.post/parent1",
		ParentCID: "bafyparent",
	}
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{Text: "reply text", Reply: reply})
	if err != nil {
		t.Fatalf("HandleCreate() err = %v", err)
	}
	if writer.reply == nil || *writer.reply != *reply {
		t.Fatalf("reply = %#v, want %#v", writer.reply, reply)
	}
}

func TestHandler_CreatePost_IncompleteReplyRejected(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{
		Text:  "reply text",
		Reply: &intent.ReplyTo{ParentURI: "at://did:plc:parent/app.bsky.feed.post/parent1"},
	})
	if err == nil {
		t.Fatal("HandleCreate() err = nil, want validation error")
	}
	if writer.text != "" {
		t.Fatalf("writer called with %q, want no call", writer.text)
	}
}
