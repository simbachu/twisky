package post

import (
	"context"
	"sync"
	"time"
)

// defaultCacheMaxEntries caps how many distinct keys a poll cache may retain.
// Expired entries are swept on access; when still over capacity, the oldest
// completed entry is evicted so long-lived processes cannot grow without bound.
const defaultCacheMaxEntries = 1024

type ttlCache[T any] struct {
	ttl        time.Duration
	maxEntries int

	mu      sync.Mutex
	entries map[string]*ttlCacheEntry[T]
}

type ttlCacheEntry[T any] struct {
	expiresAt time.Time
	val       T
	err       error
	done      chan struct{}
}

func (e *ttlCacheEntry[T]) isFresh() bool {
	select {
	case <-e.done:
		return time.Now().Before(e.expiresAt)
	default:
		return true
	}
}

func newTTLCache[T any](ttl time.Duration, maxEntries int) *ttlCache[T] {
	if maxEntries <= 0 {
		maxEntries = defaultCacheMaxEntries
	}
	return &ttlCache[T]{
		ttl:        ttl,
		maxEntries: maxEntries,
		entries:    make(map[string]*ttlCacheEntry[T]),
	}
}

// Get returns the cached value for key if it's still fresh. On a cache miss it
// calls fetch exactly once and shares the result with any other callers that
// arrive for the same key while the fetch is in flight.
//
// fetch runs under context.WithoutCancel so a cancelled owner request cannot
// poison the shared entry for waiters that are still live.
func (c *ttlCache[T]) Get(ctx context.Context, key string, fetch func(context.Context) (T, error)) (T, error) {
	c.mu.Lock()
	c.sweepLocked(time.Now())
	if entry, ok := c.entries[key]; ok {
		if entry.isFresh() {
			c.mu.Unlock()
			return waitForTTL(ctx, entry)
		}
		delete(c.entries, key)
	}

	entry := &ttlCacheEntry[T]{done: make(chan struct{})}
	c.entries[key] = entry
	c.evictOldestLocked()
	c.mu.Unlock()

	entry.val, entry.err = fetch(context.WithoutCancel(ctx))
	entry.expiresAt = time.Now().Add(c.ttl)
	close(entry.done)

	if err := ctx.Err(); err != nil {
		var zero T
		return zero, err
	}
	return entry.val, entry.err
}

func waitForTTL[T any](ctx context.Context, entry *ttlCacheEntry[T]) (T, error) {
	select {
	case <-entry.done:
		return entry.val, entry.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

func (c *ttlCache[T]) sweepLocked(now time.Time) {
	for key, entry := range c.entries {
		select {
		case <-entry.done:
			if !now.Before(entry.expiresAt) {
				delete(c.entries, key)
			}
		default:
		}
	}
}

// evictOldestLocked removes completed entries until under maxEntries.
// In-flight entries are never evicted so coalescing stays correct.
func (c *ttlCache[T]) evictOldestLocked() {
	for len(c.entries) > c.maxEntries {
		var oldestKey string
		var oldestExpiry time.Time
		found := false
		for key, entry := range c.entries {
			select {
			case <-entry.done:
				if !found || entry.expiresAt.Before(oldestExpiry) {
					oldestKey = key
					oldestExpiry = entry.expiresAt
					found = true
				}
			default:
			}
		}
		if !found {
			return
		}
		delete(c.entries, oldestKey)
	}
}

func (c *ttlCache[T]) lenForTest() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}
