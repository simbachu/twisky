package settings_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query/settings"
	"github.com/simbachu/twisky/internal/response"
)

type stubReader struct {
	profile    *bluesky.Profile
	prefs      bluesky.Preferences
	feeds      []bluesky.FeedGenerator
	err        error
	profileErr error
	feedsErr   error
}

func (s *stubReader) GetProfile(context.Context, string) (*bluesky.Profile, error) {
	if s.profileErr != nil {
		return nil, s.profileErr
	}
	return s.profile, nil
}

func (s *stubReader) GetPreferences(context.Context) (bluesky.Preferences, error) {
	if s.err != nil {
		return bluesky.Preferences{}, s.err
	}
	return s.prefs, nil
}

func (s *stubReader) GetFeedGenerators(context.Context, []string) ([]bluesky.FeedGenerator, error) {
	if s.feedsErr != nil {
		return nil, s.feedsErr
	}
	return s.feeds, nil
}

func TestHandler_HandleLoadsProfilePrefsAndFeedNames(t *testing.T) {
	t.Parallel()

	prefs, err := bluesky.ParsePreferences(nil)
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}
	prefs.AdultContentEnabled = true
	prefs.Labels["porn"] = bluesky.LabelWarn
	prefs.SavedFeeds = []bluesky.SavedFeed{{
		ID: "one", Pinned: true, Type: "feed",
		URI: "at://did:plc:feeds/app.bsky.feed.generator/for-you",
	}}
	prefs.ThreadView = bluesky.ThreadViewPref{Sort: bluesky.ThreadSortNewest, PrioritizeFollowedUsers: true}

	h := settings.NewHandler(&stubReader{
		profile: &bluesky.Profile{
			DID: "did:plc:alice", Handle: "alice.test", DisplayName: "Alice", Avatar: "https://cdn.test/a.png",
		},
		prefs: prefs,
		feeds: []bluesky.FeedGenerator{{
			URI: "at://did:plc:feeds/app.bsky.feed.generator/for-you", DisplayName: "For You",
		}},
	})

	resp := h.Handle(context.Background(), intent.ViewSettings{DID: "did:plc:alice", Handle: "alice.test"})
	view, ok := resp.(settings.SettingsView)
	if !ok {
		t.Fatalf("Handle() = %T, want SettingsView", resp)
	}
	if view.DisplayName != "Alice" || view.Handle != "alice.test" || view.Avatar == "" {
		t.Fatalf("profile = %#v", view)
	}
	if !view.AdultContentEnabled {
		t.Fatal("AdultContentEnabled = false, want true")
	}
	if len(view.Labels) != 4 || view.Labels[0].Identifier != "porn" || view.Labels[0].Value != bluesky.LabelWarn {
		t.Fatalf("Labels = %#v", view.Labels)
	}
	if len(view.Feeds) != 1 || view.Feeds[0].Label != "For You" || !view.Feeds[0].Pinned {
		t.Fatalf("Feeds = %#v", view.Feeds)
	}
	if view.ThreadSort != bluesky.ThreadSortNewest || !view.PrioritizeFollowedUsers {
		t.Fatalf("thread = sort=%q prioritize=%v", view.ThreadSort, view.PrioritizeFollowedUsers)
	}
}

func TestHandler_HandlePrefsError(t *testing.T) {
	t.Parallel()

	h := settings.NewHandler(&stubReader{err: errors.New("pds down")})
	resp := h.Handle(context.Background(), intent.ViewSettings{DID: "did:plc:alice"})
	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadGateway {
		t.Fatalf("Status = %d, want 502", errResp.Status)
	}
}

func TestHandler_HandleProfileErrorStillServesPrefs(t *testing.T) {
	t.Parallel()

	prefs, err := bluesky.ParsePreferences(nil)
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}
	h := settings.NewHandler(&stubReader{
		prefs:      prefs,
		profileErr: errors.New("appview down"),
	})

	resp := h.Handle(context.Background(), intent.ViewSettings{DID: "did:plc:alice", Handle: "alice.test"})
	view, ok := resp.(settings.SettingsView)
	if !ok {
		t.Fatalf("Handle() = %T, want SettingsView", resp)
	}
	if view.Handle != "alice.test" || view.DID != "did:plc:alice" {
		t.Fatalf("view = %#v, want session identity fallback", view)
	}
}
