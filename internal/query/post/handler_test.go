package post_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/response"
)

type stubReader struct {
	profile     *bluesky.Profile
	thread      bluesky.ThreadNode
	profiles    []bluesky.Profile
	err         error
	threadErr   error
	capturedURI string

	posts         []bluesky.Post
	postsErr      error
	postsDelay    time.Duration
	getPostsCalls int32

	mu           sync.Mutex
	capturedURIs [][]string
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

func (s *stubReader) GetPosts(_ context.Context, uris []string) ([]bluesky.Post, error) {
	atomic.AddInt32(&s.getPostsCalls, 1)
	s.mu.Lock()
	s.capturedURIs = append(s.capturedURIs, uris)
	s.mu.Unlock()
	if s.postsDelay > 0 {
		time.Sleep(s.postsDelay)
	}
	if s.postsErr != nil {
		return nil, s.postsErr
	}
	return s.posts, nil
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

	handler := post.NewHandler(reader, nil)
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
	if !view.HasAncestors {
		t.Fatal("HasAncestors = false, want true")
	}
	if len(view.Ancestors) != 0 {
		t.Fatalf("view.Ancestors = %#v, want empty on full page", view.Ancestors)
	}
	if len(view.Replies) != 1 || view.Replies[0].Post.ID != "reply1" {
		t.Fatalf("view.Replies = %#v, want one reply", view.Replies)
	}
}

func TestHandler_Handle_AncestorsFragment(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{
			DID:    "did:plc:example",
			Handle: "bsky.app",
		},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app"},
				Record: bluesky.PostRecord{Text: "root post"},
			},
			Parent: bluesky.ThreadViewPost{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/parent",
					Author: bluesky.Author{Handle: "bsky.app"},
					Record: bluesky.PostRecord{Text: "parent post"},
				},
			},
		},
	}

	handler := post.NewHandler(reader, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{
		Slug: "bsky.app",
		ID:   "root",
		Part: feedquery.PostPagePartAncestors,
	})

	view, ok := resp.(feedquery.PostPageView)
	if !ok {
		t.Fatalf("response type = %T, want PostPageView", resp)
	}
	if len(view.Ancestors) != 1 || view.Ancestors[0].Post.ID != "parent" {
		t.Fatalf("view.Ancestors = %#v, want one parent", view.Ancestors)
	}
	if view.Post.ID != "" {
		t.Fatalf("view.Post.ID = %q, want empty on ancestors fragment", view.Post.ID)
	}
}

func TestHandler_Handle_InvalidSlug(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{}, nil)
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
	}, nil)
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
	}, nil)
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
	}, nil)
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
	}, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "missing"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusNotFound)
	}
}

func TestHandler_Handle_CountsFragment_UsesGetPostsNotThread(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		posts: []bluesky.Post{
			{
				URI:         "at://did:plc:example/app.bsky.feed.post/root",
				Author:      bluesky.Author{Handle: "bsky.app"},
				Record:      bluesky.PostRecord{Text: "root post"},
				LikeCount:   42,
				RepostCount: 3,
				ReplyCount:  1,
			},
		},
		// If the handler falls back to GetPostThread this stays unset and the
		// thread-based assertions below would fail instead.
		threadErr: errors.New("counts fragment must not call GetPostThread"),
	}

	handler := post.NewHandler(reader, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{
		Slug: "bsky.app",
		ID:   "root",
		Part: feedquery.PostPagePartCounts,
	})

	view, ok := resp.(feedquery.PostPageView)
	if !ok {
		t.Fatalf("response type = %T, want PostPageView", resp)
	}
	if view.Post.LikeCount != 42 || view.Post.RepostCount != 3 || view.Post.ReplyCount != 1 {
		t.Fatalf("view.Post = %#v, want fresh counts from GetPosts", view.Post)
	}
	if reader.capturedURI != "" {
		t.Fatalf("capturedURI = %q, want GetPostThread never called", reader.capturedURI)
	}
	if len(reader.capturedURIs) != 1 || reader.capturedURIs[0][0] != "at://did:plc:example/app.bsky.feed.post/root" {
		t.Fatalf("capturedURIs = %#v, want single call with the post URI", reader.capturedURIs)
	}
}

func TestHandler_Handle_CountsFragment_NotFound(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		posts:   nil,
	}, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{
		Slug: "bsky.app",
		ID:   "missing",
		Part: feedquery.PostPagePartCounts,
	})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusNotFound)
	}
}

func TestHandler_Handle_CountsFragment_UpstreamError(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile:  &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		postsErr: errors.New("network failure"),
	}, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{
		Slug: "bsky.app",
		ID:   "root",
		Part: feedquery.PostPagePartCounts,
	})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusBadGateway)
	}
}

func TestHandler_Handle_CountsFragment_CoalescesConcurrentRequests(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		posts: []bluesky.Post{
			{URI: "at://did:plc:example/app.bsky.feed.post/root", Author: bluesky.Author{Handle: "bsky.app"}},
		},
		postsDelay: 20 * time.Millisecond,
	}
	handler := post.NewHandler(reader, nil)

	const concurrency = 10
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			resp := handler.Handle(context.Background(), intent.ViewPost{
				Slug: "bsky.app",
				ID:   "root",
				Part: feedquery.PostPagePartCounts,
			})
			if _, ok := resp.(feedquery.PostPageView); !ok {
				t.Errorf("response type = %T, want PostPageView", resp)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&reader.getPostsCalls); got != 1 {
		t.Fatalf("GetPosts calls = %d, want 1 (concurrent requests should coalesce)", got)
	}
}
