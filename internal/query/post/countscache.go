package post

import (
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
)

func newCountsCache(ttl time.Duration) *ttlCache[bluesky.Post] {
	return newTTLCache[bluesky.Post](ttl, defaultCacheMaxEntries)
}
