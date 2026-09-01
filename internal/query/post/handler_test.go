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

	threadDelay        time.Duration
	getPostThreadCalls int32
	lastThreadOpts     *bluesky.PostThreadOpts
	getProfileCalls    int32

	mu           sync.Mutex
	capturedURIs [][]string
}

func (s *stubReader) GetProfile(context.Context, string) (*bluesky.Profile, error) {
	atomic.AddInt32(&s.getProfileCalls, 1)
	return s.profile, s.err
}

func (s *stubReader) GetPostThread(_ context.Context, postURI string, opts *bluesky.PostThreadOpts) (bluesky.ThreadNode, error) {
	atomic.AddInt32(&s.getPostThreadCalls, 1)
	s.capturedURI = postURI
	s.mu.Lock()
	s.lastThreadOpts = opts
	s.mu.Unlock()
	if s.threadDelay > 0 {
		time.Sleep(s.threadDelay)
	}
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

	handler := post.NewHandler(reader, nil, nil)
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

	handler := post.NewHandler(reader, nil, nil)
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

	handler := post.NewHandler(&stubReader{}, nil, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "hello", ID: "abc"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusBadRequest)
	}
	if errResp.Message != "Invalid handle or DID" {
		t.Fatalf("message = %q, want Invalid handle or DID", errResp.Message)
	}
}

func TestHandler_Handle_InvalidPostID(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
	}, nil, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "  "})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusBadRequest)
	}
	if errResp.Message != "Invalid post identifier" {
		t.Fatalf("message = %q, want Invalid post identifier", errResp.Message)
	}
}

func TestHandler_Handle_ResolveHandleNotFound(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{err: bluesky.ErrNotFound}, nil, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "missing.example", ID: "abc"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusNotFound)
	}
	if errResp.Message != "Could not find handle missing.example" {
		t.Fatalf("message = %q, want Could not find handle missing.example", errResp.Message)
	}
}

func TestHandler_Handle_ResolveHandleUpstreamError(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{err: errors.New("network failure")}, nil, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "abc"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusBadGateway)
	}
	if errResp.Message != "Failed to resolve handle bsky.app" {
		t.Fatalf("message = %q, want Failed to resolve handle bsky.app", errResp.Message)
	}
}

func TestHandler_Handle_PostNotFound(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile:   &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		threadErr: bluesky.ErrNotFound,
	}, nil, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "missing"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusNotFound)
	}
	if errResp.Message != "Could not find post missing" {
		t.Fatalf("message = %q, want Could not find post missing", errResp.Message)
	}
}

func TestHandler_Handle_UpstreamError(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile:   &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		threadErr: errors.New("network failure"),
	}, nil, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "abc"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusBadGateway)
	}
	if errResp.Message != "Failed to load post abc" {
		t.Fatalf("message = %q, want Failed to load post abc", errResp.Message)
	}
}

func TestHandler_Handle_RootNotThreadViewPost(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread:  bluesky.NotFoundPost{URI: "at://did:plc:example/app.bsky.feed.post/missing"},
	}, nil, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "missing"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("response type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", errResp.Status, http.StatusNotFound)
	}
	if errResp.Message != "Could not find post missing" {
		t.Fatalf("message = %q, want Could not find post missing", errResp.Message)
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

	handler := post.NewHandler(reader, nil, nil)
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
	}, nil, nil)
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
	if errResp.Message != "Could not find post missing" {
		t.Fatalf("message = %q, want Could not find post missing", errResp.Message)
	}
}

func TestHandler_Handle_CountsFragment_UpstreamError(t *testing.T) {
	t.Parallel()

	handler := post.NewHandler(&stubReader{
		profile:  &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		postsErr: errors.New("network failure"),
	}, nil, nil)
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
	if errResp.Message != "Failed to refresh counts for post root" {
		t.Fatalf("message = %q, want Failed to refresh counts for post root", errResp.Message)
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
	handler := post.NewHandler(reader, nil, nil)

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

func TestHandler_Handle_RepliesFragment_UsesGetPostThread(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app"},
				Record: bluesky.PostRecord{Text: "root post"},
			},
			Replies: []bluesky.ThreadNode{
				bluesky.ThreadViewPost{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/reply1",
						Author: bluesky.Author{Handle: "dev.example"},
						Record: bluesky.PostRecord{Text: "reply one"},
					},
				},
			},
		},
		postsErr: errors.New("replies fragment must not call GetPosts"),
	}

	handler := post.NewHandler(reader, nil, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{
		Slug: "bsky.app",
		ID:   "root",
		Part: feedquery.PostPagePartReplies,
	})

	view, ok := resp.(feedquery.PostPageView)
	if !ok {
		t.Fatalf("response type = %T, want PostPageView", resp)
	}
	if view.Post.ID != "root" {
		t.Fatalf("view.Post.ID = %q, want root", view.Post.ID)
	}
	if len(view.Replies) != 1 || view.Replies[0].Post.ID != "reply1" {
		t.Fatalf("view.Replies = %#v, want reply1", view.Replies)
	}
	if reader.capturedURI != "at://did:plc:example/app.bsky.feed.post/root" {
		t.Fatalf("capturedURI = %q, want GetPostThread for the root", reader.capturedURI)
	}
	if atomic.LoadInt32(&reader.getPostsCalls) != 0 {
		t.Fatalf("GetPosts calls = %d, want 0", reader.getPostsCalls)
	}
}

func TestHandler_Handle_RepliesFragment_CoalescesConcurrentRequests(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app"},
				Record: bluesky.PostRecord{Text: "root post"},
			},
		},
		threadDelay: 20 * time.Millisecond,
	}
	handler := post.NewHandler(reader, nil, nil)

	const concurrency = 10
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			resp := handler.Handle(context.Background(), intent.ViewPost{
				Slug: "bsky.app",
				ID:   "root",
				Part: feedquery.PostPagePartReplies,
			})
			if _, ok := resp.(feedquery.PostPageView); !ok {
				t.Errorf("response type = %T, want PostPageView", resp)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&reader.getPostThreadCalls); got != 1 {
		t.Fatalf("GetPostThread calls = %d, want 1 (concurrent requests should coalesce)", got)
	}
}

func TestHandler_Handle_CountsFragment_DIDSlugSkipsGetProfile(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		err: errors.New("GetProfile must not be called for DID slug counts"),
		posts: []bluesky.Post{
			{
				URI:        "at://did:plc:example/app.bsky.feed.post/root",
				Author:     bluesky.Author{Handle: "bsky.app"},
				LikeCount:  7,
				ReplyCount: 1,
			},
		},
	}
	handler := post.NewHandler(reader, nil, nil)
	resp := handler.Handle(context.Background(), intent.ViewPost{
		Slug: "did:plc:example",
		ID:   "root",
		Part: feedquery.PostPagePartCounts,
	})
	view, ok := resp.(feedquery.PostPageView)
	if !ok {
		t.Fatalf("response type = %T, want PostPageView", resp)
	}
	if view.Post.LikeCount != 7 {
		t.Fatalf("LikeCount = %d, want 7", view.Post.LikeCount)
	}
	if atomic.LoadInt32(&reader.getProfileCalls) != 0 {
		t.Fatalf("GetProfile calls = %d, want 0", reader.getProfileCalls)
	}
}

func TestHandler_Handle_FullAndAncestorsShareThreadCache(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
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
	handler := post.NewHandler(reader, nil, nil)

	full := handler.Handle(context.Background(), intent.ViewPost{Slug: "bsky.app", ID: "root"})
	if _, ok := full.(feedquery.PostPageView); !ok {
		t.Fatalf("full response type = %T, want PostPageView", full)
	}
	ancestors := handler.Handle(context.Background(), intent.ViewPost{
		Slug: "bsky.app",
		ID:   "root",
		Part: feedquery.PostPagePartAncestors,
	})
	view, ok := ancestors.(feedquery.PostPageView)
	if !ok {
		t.Fatalf("ancestors response type = %T, want PostPageView", ancestors)
	}
	if len(view.Ancestors) != 1 || view.Ancestors[0].Post.ID != "parent" {
		t.Fatalf("ancestors = %#v, want parent", view.Ancestors)
	}
	if got := atomic.LoadInt32(&reader.getPostThreadCalls); got != 1 {
		t.Fatalf("GetPostThread calls = %d, want 1 (full+ancestors share cache)", got)
	}
	if reader.lastThreadOpts == nil || reader.lastThreadOpts.Depth != 6 || reader.lastThreadOpts.ParentHeight != 25 {
		t.Fatalf("lastThreadOpts = %#v, want depth=6 parentHeight=25", reader.lastThreadOpts)
	}
}

func TestHandler_WithReader_SharesThreadCache(t *testing.T) {
	t.Parallel()

	thread := bluesky.ThreadViewPost{
		Post: bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/root",
			Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
			Record: bluesky.PostRecord{Text: "root post"},
		},
	}
	public := &stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread:  thread,
	}
	auth := &stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread:  thread,
	}

	handler := post.NewHandler(public, nil, nil)
	intent := intent.ViewPost{Slug: "bsky.app", ID: "root"}

	if _, ok := handler.WithReader(auth).Handle(context.Background(), intent).(feedquery.PostPageView); !ok {
		t.Fatal("first Handle() want PostPageView")
	}
	if _, ok := handler.WithReader(auth).Handle(context.Background(), intent).(feedquery.PostPageView); !ok {
		t.Fatal("second Handle() want PostPageView")
	}
	if got := atomic.LoadInt32(&public.getPostThreadCalls); got != 0 {
		t.Fatalf("public GetPostThread calls = %d, want 0", got)
	}
	if got := atomic.LoadInt32(&auth.getPostThreadCalls); got != 1 {
		t.Fatalf("auth GetPostThread calls = %d, want 1 (shared cache)", got)
	}
}

func TestHandler_HandleThread_OPThoughtSpine(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{DID: "did:plc:alice", Handle: "alice.test"},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:alice/app.bsky.feed.post/p1",
				Author: bluesky.Author{DID: "did:plc:alice", Handle: "alice.test"},
				Record: bluesky.PostRecord{Text: "one"},
			},
			Replies: []bluesky.ThreadNode{
				bluesky.ThreadViewPost{
					Post: bluesky.Post{
						URI:    "at://did:plc:bob/app.bsky.feed.post/r1",
						Author: bluesky.Author{DID: "did:plc:bob", Handle: "bob.test"},
						Record: bluesky.PostRecord{Text: "bob"},
					},
				},
				bluesky.ThreadViewPost{
					Post: bluesky.Post{
						URI:    "at://did:plc:alice/app.bsky.feed.post/p2",
						Author: bluesky.Author{DID: "did:plc:alice", Handle: "alice.test"},
						Record: bluesky.PostRecord{Text: "two"},
					},
				},
			},
		},
	}
	handler := post.NewHandler(reader, nil, nil)

	resp := handler.HandleThread(context.Background(), intent.ViewThread{Slug: "alice.test", ID: "p1"})
	view, ok := resp.(feedquery.ThoughtThreadView)
	if !ok {
		t.Fatalf("response type = %T, want ThoughtThreadView", resp)
	}
	if len(view.Posts) != 2 {
		t.Fatalf("len(Posts) = %d, want 2", len(view.Posts))
	}
	if view.Posts[0].ID != "p1" || view.Posts[1].ID != "p2" {
		t.Fatalf("posts = %q %q, want p1 p2", view.Posts[0].ID, view.Posts[1].ID)
	}
	if view.RootID != "p1" {
		t.Fatalf("RootID = %q, want p1", view.RootID)
	}
}
