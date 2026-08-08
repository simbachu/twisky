package post

import (
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
)

func newThreadCache(ttl time.Duration) *ttlCache[bluesky.ThreadNode] {
	return newTTLCache[bluesky.ThreadNode](ttl, defaultCacheMaxEntries)
}
