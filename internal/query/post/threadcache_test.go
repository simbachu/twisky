package post

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
)

func TestThreadCache_CoalescesConcurrentFetches(t *testing.T) {
	t.Parallel()

	cache := newThreadCache(time.Minute)
	var calls int32
	fetch := func(context.Context) (bluesky.ThreadNode, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(10 * time.Millisecond)
		return bluesky.ThreadViewPost{
			Post: bluesky.Post{URI: "at://example/app.bsky.feed.post/abc"},
		}, nil
	}

	const concurrency = 8
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			if _, err := cache.Get(context.Background(), "key", fetch); err != nil {
				t.Errorf("Get() err = %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("fetch calls = %d, want 1", got)
	}
}

func TestThreadCache_ReusesResultWithinTTL(t *testing.T) {
	t.Parallel()

	cache := newThreadCache(time.Minute)
	var calls int32
	fetch := func(context.Context) (bluesky.ThreadNode, error) {
		atomic.AddInt32(&calls, 1)
		return bluesky.ThreadViewPost{
			Post: bluesky.Post{URI: "at://example/app.bsky.feed.post/abc"},
		}, nil
	}

	if _, err := cache.Get(context.Background(), "key", fetch); err != nil {
		t.Fatalf("Get() err = %v", err)
	}
	if _, err := cache.Get(context.Background(), "key", fetch); err != nil {
		t.Fatalf("Get() err = %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("fetch calls = %d, want 1 (second call should hit cache)", got)
	}
}

func TestThreadCache_RefetchesAfterTTLExpires(t *testing.T) {
	t.Parallel()

	cache := newThreadCache(10 * time.Millisecond)
	var calls int32
	fetch := func(context.Context) (bluesky.ThreadNode, error) {
		atomic.AddInt32(&calls, 1)
		return bluesky.ThreadViewPost{
			Post: bluesky.Post{URI: "at://example/app.bsky.feed.post/abc"},
		}, nil
	}

	if _, err := cache.Get(context.Background(), "key", fetch); err != nil {
		t.Fatalf("Get() err = %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	if _, err := cache.Get(context.Background(), "key", fetch); err != nil {
		t.Fatalf("Get() err = %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("fetch calls = %d, want 2 (entry should have expired)", got)
	}
}
