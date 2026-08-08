package http

import (
	"net/http"
	"sync"
	"time"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/auth/session"
)

// sessionClientCacheTTL reuses a resumed API client across nearby authenticated
// poll/enrichment requests so ResumeSession does not hit SQLite on every tick.
const sessionClientCacheTTL = 30 * time.Second

type sessionClientCacheEntry struct {
	account   *session.Account
	client    *authoauth.SessionClient
	expiresAt time.Time
}

type sessionClientCache struct {
	mu      sync.Mutex
	entries map[string]sessionClientCacheEntry
}

func newSessionClientCache() *sessionClientCache {
	return &sessionClientCache{entries: make(map[string]sessionClientCacheEntry)}
}

func (c *sessionClientCache) get(key string) (sessionClientCacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok || !time.Now().Before(entry.expiresAt) {
		if ok {
			delete(c.entries, key)
		}
		return sessionClientCacheEntry{}, false
	}
	return entry, true
}

func (c *sessionClientCache) put(key string, account *session.Account, client *authoauth.SessionClient) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = sessionClientCacheEntry{
		account:   account,
		client:    client,
		expiresAt: time.Now().Add(sessionClientCacheTTL),
	}
	// Bound map growth: drop expired keys opportunistically.
	now := time.Now()
	for k, entry := range c.entries {
		if !now.Before(entry.expiresAt) {
			delete(c.entries, k)
		}
	}
}

func sessionClientCacheKey(account *session.Account) string {
	return account.DID + "/" + account.SessionID
}

// cachedResumeActiveClient returns a short-lived cached SessionClient when available.
func (s *Server) cachedResumeActiveClient(r *http.Request) (*session.Account, *authoauth.SessionClient, error) {
	account, err := s.loadActiveAccount(r)
	if err != nil {
		return nil, nil, err
	}
	if s.sessionClients == nil {
		return s.resumeActiveClientUncached(r, account)
	}
	key := sessionClientCacheKey(account)
	if entry, ok := s.sessionClients.get(key); ok {
		return entry.account, entry.client, nil
	}
	account, client, err := s.resumeActiveClientUncached(r, account)
	if err != nil {
		return nil, nil, err
	}
	s.sessionClients.put(key, account, client)
	return account, client, nil
}
