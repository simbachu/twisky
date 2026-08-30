package settings

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/response"
)

type Reader interface {
	GetProfile(ctx context.Context, actor string) (*bluesky.Profile, error)
	GetPreferences(ctx context.Context) (bluesky.Preferences, error)
	GetFeedGenerators(ctx context.Context, uris []string) ([]bluesky.FeedGenerator, error)
}

type Handler struct {
	reader Reader
}

func NewHandler(reader Reader) *Handler {
	return &Handler{reader: reader}
}

type LabelSetting struct {
	Identifier string
	Name       string
	Value      bluesky.LabelVisibility
}

type FeedSetting struct {
	Label  string
	URI    string
	Pinned bool
	Type   string
}

// SettingsView is the read model for the settings page.
type SettingsView struct {
	DID                     string
	Handle                  string
	DisplayName             string
	Avatar                  string
	AdultContentEnabled     bool
	Labels                  []LabelSetting
	Feeds                   []FeedSetting
	ThreadSort              string
	PrioritizeFollowedUsers bool
}

func (SettingsView) IsResponse() {}

// ViewFromPreferences builds the settings read model from parsed Bluesky prefs.
func ViewFromPreferences(prefs bluesky.Preferences) SettingsView {
	return SettingsView{
		AdultContentEnabled:     prefs.AdultContentEnabled,
		Labels:                  labelSettings(prefs),
		Feeds:                   feedSettings(prefs.SavedFeeds),
		ThreadSort:              prefs.ThreadView.Sort,
		PrioritizeFollowedUsers: prefs.ThreadView.PrioritizeFollowedUsers,
	}
}

var labelNames = map[string]string{
	"porn":          "Pornography",
	"sexual":        "Suggestive content",
	"nudity":        "Non-sexual nudity",
	"graphic-media": "Graphic media",
}

func (h *Handler) Handle(ctx context.Context, i intent.ViewSettings) response.Response {
	if h.reader == nil {
		return response.ErrorResponse{Status: http.StatusInternalServerError, Message: "Something went wrong loading this page"}
	}
	prefs, err := h.reader.GetPreferences(ctx)
	if err != nil {
		slog.Error("settings preferences unavailable", "err", err)
		if authoauth.IsInsufficientAuth(err) {
			return response.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Your Bluesky login needs to be refreshed. Log out and sign in again to load settings.",
			}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "Failed to load settings"}
	}

	view := ViewFromPreferences(prefs)
	view.DID = i.DID
	view.Handle = i.Handle
	view.DisplayName = i.Handle

	actor := i.DID
	if actor == "" {
		actor = i.Handle
	}
	if actor != "" {
		profile, err := h.reader.GetProfile(ctx, actor)
		if err != nil {
			if !errors.Is(err, bluesky.ErrNotFound) {
				slog.Warn("settings profile unavailable", "err", err)
			}
		} else if profile != nil {
			view.DID = profile.DID
			view.Handle = profile.Handle
			view.DisplayName = profile.DisplayName
			view.Avatar = profile.Avatar
			if view.DisplayName == "" {
				view.DisplayName = profile.Handle
			}
		}
	}

	view.Feeds = h.resolveFeedLabels(ctx, view.Feeds)
	return view
}

func labelSettings(prefs bluesky.Preferences) []LabelSetting {
	out := make([]LabelSetting, 0, len(bluesky.ContentFilterLabels))
	for _, identifier := range bluesky.ContentFilterLabels {
		name := labelNames[identifier]
		if name == "" {
			name = identifier
		}
		value := bluesky.LabelIgnore
		if prefs.Labels != nil {
			if vis, ok := prefs.Labels[identifier]; ok {
				value = vis
			}
		}
		out = append(out, LabelSetting{
			Identifier: identifier,
			Name:       name,
			Value:      value,
		})
	}
	return out
}

func feedSettings(feeds []bluesky.SavedFeed) []FeedSetting {
	out := make([]FeedSetting, 0, len(feeds))
	for _, feed := range feeds {
		label := feed.URI
		if label == "" {
			label = feed.ID
		}
		out = append(out, FeedSetting{
			Label:  label,
			URI:    feed.URI,
			Pinned: feed.Pinned,
			Type:   feed.Type,
		})
	}
	return out
}

func (h *Handler) resolveFeedLabels(ctx context.Context, feeds []FeedSetting) []FeedSetting {
	uris := make([]string, 0, len(feeds))
	for _, feed := range feeds {
		if feed.Type == "feed" && feed.URI != "" {
			uris = append(uris, feed.URI)
		}
	}
	if len(uris) == 0 {
		return feeds
	}
	generators, err := h.reader.GetFeedGenerators(ctx, uris)
	if err != nil {
		slog.Warn("settings feed generators unavailable", "err", err)
		return feeds
	}
	byURI := make(map[string]string, len(generators))
	for _, generator := range generators {
		if generator.DisplayName != "" {
			byURI[generator.URI] = generator.DisplayName
		}
	}
	for i, feed := range feeds {
		if name, ok := byURI[feed.URI]; ok {
			feeds[i].Label = name
		}
	}
	return feeds
}
