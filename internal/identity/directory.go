package identity

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/bluesky"
)

var (
	ErrNotFound           = bluesky.ErrNotFound
	ErrResolutionFailed   = errors.New("identity resolution failed")
)

const invalidHandle = "handle.invalid"

// ProfileReader resolves an actor via AppView when indigo lookup misses.
type ProfileReader interface {
	GetProfile(ctx context.Context, actor string) (*bluesky.Profile, error)
}

// Directory wraps indigo identity resolution with a process-wide last-known-good
// handle↔DID map so transient handle verification failures do not take pages down.
type Directory struct {
	inner identity.Directory

	mu          sync.RWMutex
	handleToDID map[string]string
	didToHandle map[string]string
}

// NewDirectory returns a long-lived identity directory for the process.
func NewDirectory() *Directory {
	return NewDirectoryWithInner(newConfiguredInner())
}

// NewDirectoryWithInner wraps a custom indigo Directory (tests).
func NewDirectoryWithInner(inner identity.Directory) *Directory {
	return &Directory{
		inner:       inner,
		handleToDID: make(map[string]string),
		didToHandle: make(map[string]string),
	}
}

func newConfiguredInner() identity.Directory {
	base := identity.BaseDirectory{
		PLCURL: identity.DefaultPLCURL,
		HTTPClient: http.Client{
			Timeout: 2 * time.Second,
			Transport: &http.Transport{
				Proxy:             http.ProxyFromEnvironment,
				IdleConnTimeout:   time.Second,
				MaxIdleConns:      100,
			},
		},
		Resolver: net.Resolver{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 3 * time.Second}
				return d.DialContext(ctx, network, address)
			},
		},
		TryAuthoritativeDNS:   true,
		FallbackDNSServers:    []string{"1.1.1.1:53", "8.8.8.8:53"},
		SkipDNSDomainSuffixes: []string{".bsky.social"},
		UserAgent:             "twisky-identity/" + versioninfo.Short(),
	}
	inner := identity.NewCacheDirectory(&base, 250_000, 24*time.Hour, 2*time.Minute, 5*time.Minute)
	return inner
}

// Inner returns the indigo Directory shared with OAuth ClientApp.
func (d *Directory) Inner() identity.Directory {
	if d == nil {
		return identity.DefaultDirectory()
	}
	return d.inner
}

func (d *Directory) LookupHandle(ctx context.Context, h syntax.Handle) (*identity.Identity, error) {
	ident, err := d.inner.LookupHandle(ctx, h)
	if err == nil && ident != nil {
		d.Observe(h.String(), ident.DID.String())
	}
	return ident, err
}

func (d *Directory) LookupDID(ctx context.Context, did syntax.DID) (*identity.Identity, error) {
	ident, err := d.inner.LookupDID(ctx, did)
	if err == nil && ident != nil && !ident.Handle.IsInvalidHandle() {
		d.Observe(ident.Handle.String(), ident.DID.String())
	}
	return ident, err
}

func (d *Directory) Lookup(ctx context.Context, atid syntax.AtIdentifier) (*identity.Identity, error) {
	ident, err := d.inner.Lookup(ctx, atid)
	if err == nil && ident != nil && !ident.Handle.IsInvalidHandle() {
		d.Observe(ident.Handle.String(), ident.DID.String())
	}
	return ident, err
}

func (d *Directory) Purge(ctx context.Context, atid syntax.AtIdentifier) error {
	return d.inner.Purge(ctx, atid)
}

// Resolve returns the DID for a URL slug. DID slugs pass through unchanged.
// Resolution order: last-known-good, indigo LookupHandle, AppView getProfile.
func (d *Directory) Resolve(ctx context.Context, slug actor.Slug, reader ProfileReader) (string, error) {
	if slug.Kind == actor.KindDID {
		return slug.Identifier, nil
	}
	handle := slug.Identifier

	if did := d.cachedDID(handle); did != "" {
		return did, nil
	}

	if did, err := d.resolveLive(ctx, handle, reader); err == nil {
		return did, nil
	} else if did := d.cachedDID(handle); did != "" {
		return did, nil
	} else if err != nil {
		return "", err
	}

	return "", ErrResolutionFailed
}

func (d *Directory) resolveLive(ctx context.Context, handle string, reader ProfileReader) (string, error) {
	var liveErr error

	parsed, err := syntax.ParseHandle(handle)
	if err == nil {
		ident, lookupErr := d.inner.LookupHandle(ctx, parsed)
		if lookupErr == nil && ident != nil {
			d.Observe(handle, ident.DID.String())
			return ident.DID.String(), nil
		}
		if lookupErr != nil {
			liveErr = lookupErr
		}
	}

	if reader != nil {
		profile, profileErr := reader.GetProfile(ctx, handle)
		if profileErr == nil && profile != nil && profile.DID != "" {
			d.Observe(handle, profile.DID)
			return profile.DID, nil
		}
		if errors.Is(profileErr, bluesky.ErrNotFound) {
			return "", ErrNotFound
		}
		if profileErr != nil {
			liveErr = profileErr
		}
	}

	if liveErr != nil {
		return "", liveErr
	}
	return "", ErrResolutionFailed
}

// Observe records a verified handle↔DID mapping. Invalid handles are ignored.
func (d *Directory) Observe(handle, did string) {
	if d == nil || handle == "" || did == "" || handle == invalidHandle {
		return
	}
	if _, err := actor.ParseSlug(handle); err != nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handleToDID[handle] = did
	d.didToHandle[did] = handle
}

// ObserveProfile records mapping from an AppView profile payload.
func (d *Directory) ObserveProfile(profile *bluesky.Profile) {
	if profile == nil {
		return
	}
	d.Observe(profile.Handle, profile.DID)
}

// ObserveAuthor records mapping from an AppView author payload.
func (d *Directory) ObserveAuthor(author bluesky.Author) {
	d.Observe(author.Handle, author.DID)
}

// HandleFor returns the last known real handle for a DID, or empty.
func (d *Directory) HandleFor(did string) string {
	if d == nil || did == "" {
		return ""
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.didToHandle[did]
}

// DisplayHandle prefers a real AppView handle, then last-known-good, then the raw value.
func (d *Directory) DisplayHandle(appViewHandle, did string) string {
	if appViewHandle != "" && appViewHandle != invalidHandle {
		return appViewHandle
	}
	if h := d.HandleFor(did); h != "" {
		return h
	}
	return appViewHandle
}

func (d *Directory) cachedDID(handle string) string {
	if d == nil || handle == "" {
		return ""
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.handleToDID[handle]
}

// KnownDID returns a cached handle→DID mapping without live lookup.
func (d *Directory) KnownDID(handle string) string {
	return d.cachedDID(handle)
}
