package post

import (
	"context"
	"sync"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
)

// countsCache de-duplicates concurrent counts-only fetches for the same post
// key within a short TTL, so many browser tabs polling the same post at once
// only trigger a single upstream call.
type countsCache struct {
	ttl time.Duration

	mu      sync.Mutex
	entries map[string]*countsCacheEntry
}

type countsCacheEntry struct {
	expiresAt time.Time
	post      bluesky.Post
	err       error
	done      chan struct{}
}

// isFresh reports whether callers should reuse this entry rather than
// triggering a new fetch. An in-flight entry (fetch not yet complete) is
// always fresh, since concurrent callers should wait for it instead of
// racing a duplicate fetch; a completed entry is fresh only within the TTL.
func (e *countsCacheEntry) isFresh() bool {
	select {
	case <-e.done:
		return time.Now().Before(e.expiresAt)
	default:
		return true
	}
}

func newCountsCache(ttl time.Duration) *countsCache {
	return &countsCache{ttl: ttl, entries: make(map[string]*countsCacheEntry)}
}

// Get returns the cached post for key if it's still fresh. On a cache miss it
// calls fetch exactly once and shares the result with any other callers that
// arrive for the same key while the fetch is in flight.
func (c *countsCache) Get(ctx context.Context, key string, fetch func(context.Context) (bluesky.Post, error)) (bluesky.Post, error) {
	c.mu.Lock()
	if entry, ok := c.entries[key]; ok {
		if entry.isFresh() {
			c.mu.Unlock()
			return waitFor(ctx, entry)
		}
		delete(c.entries, key)
	}

	entry := &countsCacheEntry{done: make(chan struct{})}
	c.entries[key] = entry
	c.mu.Unlock()

	entry.post, entry.err = fetch(ctx)
	entry.expiresAt = time.Now().Add(c.ttl)
	close(entry.done)

	return entry.post, entry.err
}

func waitFor(ctx context.Context, entry *countsCacheEntry) (bluesky.Post, error) {
	select {
	case <-entry.done:
		return entry.post, entry.err
	case <-ctx.Done():
		return bluesky.Post{}, ctx.Err()
	}
}
