package post_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/response"
)

type stubReader struct {
	profile    *bluesky.Profile
	thread     bluesky.ThreadNode
	profiles   []bluesky.Profile
	err        error
	threadErr  error
	capturedURI string
}

func (s *stubReader) GetProfile(context.Context, string) (*bluesky.Profile, error) {
	return s.profile, s.err
}

func (s *stubReader) GetPostThread(_ context.Context, postURI string) (bluesky.ThreadNode, error) {
	s.capturedURI = postURI
	if s.threadErr != nil {
		return nil, s.threadErr
	}
	return s.thread, nil
}

func (s *stubReader) GetProfiles(context.Context, []string) ([]bluesky.Profile, error) {
	return s.profiles, nil
}

func TestHandler_Handle_OK(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{
			DID:    "did:plc:example",
			Handle: "bsky.app",
		},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
				Record: bluesky.PostRecord{Text: "root post"},
			},
			Parent: bluesky.ThreadViewPost{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/parent",
					Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
					Record: bluesky.PostRecord{Text: "parent post"},
				},
			},
			Replies: []bluesky.ThreadNode{
				bluesky.ThreadViewPost{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/reply1",
						Author: bluesky.Author{Handle: "dev.example", DisplayName: "Dev"},
						Record: bluesky.PostRecord{Text: "reply one"},
					},
				},
			},
		},
	}

	handler := post.NewHandler(reader)
	resp := handler.Handle(context.Background(), intent.ViewPost{
		Slug: "bsky.app",
		ID:   "root",
	})

	view, ok := resp.(feedquery.PostPageView)
	if !ok {
		t.Fatalf("response type = %T, want PostPageView", resp)
	}
	if reader.capturedURI != "at://did:plc:example/app.bsky.feed.post/root" {
		t.Fatalf("capturedURI = %q, want at://did:plc:example/app.bsky.feed.post/root", reader.capturedURI)
	}
	if view.Post.ID != "root" || view.Post.Text != "root post" {
		t.Fatalf("view.Post = %#v, want root post", view.Post)
	}
	if len(view.Ancestors) != 1 || view.Ancestors[0].ID != "parent" {
		t.Fatalf("view.Ancestors = %#v, want one parent", view.Ancestors)
	}
	if len(view.Replies) != 1 || view.Replies[0].Post.ID != "reply1" {
		t.Fatalf("view.Replies = %#v, want one reply", view.Replies)
	}
}

func TestHandler_Handle_InvalidSlug(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{})
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "hello", ID: "abc"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusBadRequest)
	}
}

func TestHandler_Handle_InvalidPostID(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
	})
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "  "})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusBadRequest)
	}
}

func TestHandler_Handle_PostNotFound(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile:   &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		threadErr: bluesky.ErrNotFound,
	})
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "missing"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusNotFound)
	}
}

func TestHandler_Handle_UpstreamError(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile:   &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		threadErr: errors.New("network failure"),
	})
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "abc"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusBadGateway)
	}
}

func TestHandler_Handle_RootNotThreadViewPost(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread:  bluesky.NotFoundPost{URI: "at://did:plc:example/app.bsky.feed.post/missing"},
	})
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "missing"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusNotFound)
	}
}
