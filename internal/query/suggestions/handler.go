package suggestions

import (
	"context"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/identity"
)

var DefaultHandles = []string{
	"simbachu.com",
	"bsky.app",
}

type Reader interface {
	GetProfile(ctx context.Context, actor string) (*bluesky.Profile, error)
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
}

type AccountView struct {
	Handle      string
	DisplayName string
	DID         string
	Avatar      string
	IsLabeler   bool
}

type Handler struct {
	reader  Reader
	handles []string
	dir     *identity.Directory
}

func NewHandler(reader Reader, handles []string, dir *identity.Directory) *Handler {
	if len(handles) == 0 {
		handles = DefaultHandles
	}
	return &Handler{reader: reader, handles: handles, dir: dir}
}

func (h *Handler) SuggestedAccounts(ctx context.Context) []AccountView {
	if h.reader == nil || len(h.handles) == 0 {
		return nil
	}

	actors := h.resolveActors(ctx)
	profiles, err := h.reader.GetProfiles(ctx, actors)
	if err != nil {
		return h.cachedAccounts()
	}

	byHandle := make(map[string]bluesky.Profile, len(profiles))
	byDID := make(map[string]bluesky.Profile, len(profiles))
	for _, profile := range profiles {
		if h.dir != nil {
			h.dir.ObserveProfile(&profile)
		}
		byHandle[profile.Handle] = profile
		byDID[profile.DID] = profile
	}

	accounts := make([]AccountView, 0, len(h.handles))
	for _, handle := range h.handles {
		profile, ok := byHandle[handle]
		if !ok && h.dir != nil {
			slug, err := actor.ParseSlug(handle)
			if err == nil {
				if did, err := h.dir.Resolve(ctx, slug, h.reader); err == nil {
					profile, ok = byDID[did]
				}
			}
		}
		if !ok {
			continue
		}
		displayHandle := profile.Handle
		if h.dir != nil {
			displayHandle = h.dir.DisplayHandle(profile.Handle, profile.DID)
		}
		accounts = append(accounts, AccountView{
			Handle:      displayHandle,
			DisplayName: actor.Name(profile.DisplayName, displayHandle),
			DID:         profile.DID,
			Avatar:      profile.Avatar,
			IsLabeler:   actor.IsLabelerAccount(displayHandle, profile.DID, profile.Associated != nil && profile.Associated.Labeler),
		})
	}
	if len(accounts) == 0 {
		return h.cachedAccounts()
	}
	return accounts
}

func (h *Handler) resolveActors(ctx context.Context) []string {
	actors := make([]string, 0, len(h.handles))
	for _, handle := range h.handles {
		slug, err := actor.ParseSlug(handle)
		if err != nil {
			continue
		}
		if h.dir != nil {
			if did, err := h.dir.Resolve(ctx, slug, h.reader); err == nil && did != "" {
				actors = append(actors, did)
				continue
			}
		}
		actors = append(actors, handle)
	}
	return actors
}

func (h *Handler) cachedAccounts() []AccountView {
	if h.dir == nil {
		return nil
	}
	accounts := make([]AccountView, 0, len(h.handles))
	for _, handle := range h.handles {
		did := h.dir.KnownDID(handle)
		if did == "" {
			continue
		}
		displayHandle := h.dir.DisplayHandle(handle, did)
		accounts = append(accounts, AccountView{
			Handle:      displayHandle,
			DisplayName: displayHandle,
			DID:         did,
		})
	}
	if len(accounts) == 0 {
		return nil
	}
	return accounts
}
