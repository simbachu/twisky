package post

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
)

func TestTTLCache_CoalescesConcurrentFetches(t *testing.T) {
	t.Parallel()

	cache := newTTLCache[bluesky.Post](time.Minute, defaultCacheMaxEntries)
	var calls int32
	fetch := func(context.Context) (bluesky.Post, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(10 * time.Millisecond)
		return bluesky.Post{URI: "at://example/app.bsky.feed.post/abc"}, nil
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

func TestTTLCache_ReusesResultWithinTTL(t *testing.T) {
	t.Parallel()

	cache := newTTLCache[bluesky.Post](time.Minute, defaultCacheMaxEntries)
	var calls int32
	fetch := func(context.Context) (bluesky.Post, error) {
		atomic.AddInt32(&calls, 1)
		return bluesky.Post{URI: "at://example/app.bsky.feed.post/abc"}, nil
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

func TestTTLCache_RefetchesAfterTTLExpires(t *testing.T) {
	t.Parallel()

	cache := newTTLCache[bluesky.Post](10*time.Millisecond, defaultCacheMaxEntries)
	var calls int32
	fetch := func(context.Context) (bluesky.Post, error) {
		atomic.AddInt32(&calls, 1)
		return bluesky.Post{URI: "at://example/app.bsky.feed.post/abc"}, nil
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

func TestTTLCache_DifferentKeysDoNotShareResult(t *testing.T) {
	t.Parallel()

	cache := newTTLCache[bluesky.Post](time.Minute, defaultCacheMaxEntries)
	var calls int32
	fetch := func(context.Context) (bluesky.Post, error) {
		atomic.AddInt32(&calls, 1)
		return bluesky.Post{URI: "at://example/app.bsky.feed.post/abc"}, nil
	}

	if _, err := cache.Get(context.Background(), "key-a", fetch); err != nil {
		t.Fatalf("Get() err = %v", err)
	}
	if _, err := cache.Get(context.Background(), "key-b", fetch); err != nil {
		t.Fatalf("Get() err = %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("fetch calls = %d, want 2 (different keys should not coalesce)", got)
	}
}

func TestTTLCache_CancelledOwnerDoesNotPoisonWaiters(t *testing.T) {
	t.Parallel()

	cache := newTTLCache[bluesky.Post](time.Minute, defaultCacheMaxEntries)
	started := make(chan struct{})
	release := make(chan struct{})
	var calls int32
	fetch := func(ctx context.Context) (bluesky.Post, error) {
		atomic.AddInt32(&calls, 1)
		close(started)
		<-release
		if err := ctx.Err(); err != nil {
			t.Errorf("fetch ctx cancelled: %v", err)
		}
		return bluesky.Post{URI: "at://example/app.bsky.feed.post/abc"}, nil
	}

	ownerCtx, cancel := context.WithCancel(context.Background())
	ownerDone := make(chan error, 1)
	go func() {
		_, err := cache.Get(ownerCtx, "key", fetch)
		ownerDone <- err
	}()

	<-started
	cancel()
	close(release)
	ownerErr := <-ownerDone
	if ownerErr == nil {
		t.Fatal("owner Get() err = nil, want context cancellation")
	}

	post, err := cache.Get(context.Background(), "key", fetch)
	if err != nil {
		t.Fatalf("waiter Get() err = %v", err)
	}
	if post.URI == "" {
		t.Fatal("waiter got empty post")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("fetch calls = %d, want 1 (owner cancel must not start a second fetch)", got)
	}
}

func TestTTLCache_EvictsWhenOverMaxEntries(t *testing.T) {
	t.Parallel()

	cache := newTTLCache[bluesky.Post](time.Minute, 2)
	fetch := func(uri string) func(context.Context) (bluesky.Post, error) {
		return func(context.Context) (bluesky.Post, error) {
			return bluesky.Post{URI: uri}, nil
		}
	}

	if _, err := cache.Get(context.Background(), "a", fetch("a")); err != nil {
		t.Fatalf("Get(a) err = %v", err)
	}
	if _, err := cache.Get(context.Background(), "b", fetch("b")); err != nil {
		t.Fatalf("Get(b) err = %v", err)
	}
	if _, err := cache.Get(context.Background(), "c", fetch("c")); err != nil {
		t.Fatalf("Get(c) err = %v", err)
	}

	if got := cache.lenForTest(); got != 2 {
		t.Fatalf("len = %d, want 2 after eviction", got)
	}
}

func TestTTLCache_ThreadNode(t *testing.T) {
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
		t.Fatalf("fetch calls = %d, want 1", got)
	}
}
