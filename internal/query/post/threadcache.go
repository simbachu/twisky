package post

import (
	"context"
	"sync"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
)

// threadCache de-duplicates concurrent GetPostThread fetches for the same post
// key within a short TTL, so concurrent reply-refresh pollers of the same post
// only trigger a single upstream call.
type threadCache struct {
	ttl time.Duration

	mu      sync.Mutex
	entries map[string]*threadCacheEntry
}

type threadCacheEntry struct {
	expiresAt time.Time
	thread    bluesky.ThreadNode
	err       error
	done      chan struct{}
}

func (e *threadCacheEntry) isFresh() bool {
	select {
	case <-e.done:
		return time.Now().Before(e.expiresAt)
	default:
		return true
	}
}

func newThreadCache(ttl time.Duration) *threadCache {
	return &threadCache{ttl: ttl, entries: make(map[string]*threadCacheEntry)}
}

// Get returns the cached thread for key if it's still fresh. On a cache miss it
// calls fetch exactly once and shares the result with any other callers that
// arrive for the same key while the fetch is in flight.
func (c *threadCache) Get(ctx context.Context, key string, fetch func(context.Context) (bluesky.ThreadNode, error)) (bluesky.ThreadNode, error) {
	c.mu.Lock()
	if entry, ok := c.entries[key]; ok {
		if entry.isFresh() {
			c.mu.Unlock()
			return waitForThread(ctx, entry)
		}
		delete(c.entries, key)
	}

	entry := &threadCacheEntry{done: make(chan struct{})}
	c.entries[key] = entry
	c.mu.Unlock()

	entry.thread, entry.err = fetch(ctx)
	entry.expiresAt = time.Now().Add(c.ttl)
	close(entry.done)

	return entry.thread, entry.err
}

func waitForThread(ctx context.Context, entry *threadCacheEntry) (bluesky.ThreadNode, error) {
	select {
	case <-entry.done:
		return entry.thread, entry.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
