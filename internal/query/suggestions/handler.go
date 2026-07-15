package suggestions

import (
	"context"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/bluesky"
)

var DefaultHandles = []string{
	"simbachu.com",
	"bsky.app",
}

type Reader interface {
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
}

type AccountView struct {
	Handle      string
	DisplayName string
	Avatar      string
}

type Handler struct {
	reader  Reader
	handles []string
}

func NewHandler(reader Reader, handles []string) *Handler {
	if len(handles) == 0 {
		handles = DefaultHandles
	}
	return &Handler{reader: reader, handles: handles}
}

func (h *Handler) SuggestedAccounts(ctx context.Context) []AccountView {
	if h.reader == nil || len(h.handles) == 0 {
		return nil
	}

	profiles, err := h.reader.GetProfiles(ctx, h.handles)
	if err != nil {
		return nil
	}

	byHandle := make(map[string]bluesky.Profile, len(profiles))
	for _, profile := range profiles {
		byHandle[profile.Handle] = profile
	}

	accounts := make([]AccountView, 0, len(h.handles))
	for _, handle := range h.handles {
		profile, ok := byHandle[handle]
		if !ok {
			continue
		}
		accounts = append(accounts, AccountView{
			Handle:      profile.Handle,
			DisplayName: actor.Name(profile.DisplayName, profile.Handle),
			Avatar:      profile.Avatar,
		})
	}
	return accounts
}
