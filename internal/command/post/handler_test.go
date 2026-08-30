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
	recordURI string
	err       error
}

func (s *stubWriter) CreatePost(_ context.Context, text string) (string, error) {
	s.text = text
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
