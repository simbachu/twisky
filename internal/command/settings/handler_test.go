package settings_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	settingscommand "github.com/simbachu/twisky/internal/command/settings"
	"github.com/simbachu/twisky/internal/intent"
)

type stubWriter struct {
	prefs    bluesky.Preferences
	put      bluesky.Preferences
	getErr   error
	putErr   error
	putCalls int
}

func (s *stubWriter) GetPreferences(context.Context) (bluesky.Preferences, error) {
	if s.getErr != nil {
		return bluesky.Preferences{}, s.getErr
	}
	return s.prefs, nil
}

func (s *stubWriter) PutPreferences(_ context.Context, prefs bluesky.Preferences) error {
	s.putCalls++
	s.put = prefs
	return s.putErr
}

func loadedPrefs(t *testing.T) bluesky.Preferences {
	t.Helper()
	prefs, err := bluesky.ParsePreferences([]json.RawMessage{
		json.RawMessage(`{"$type":"app.bsky.actor.defs#adultContentPref","enabled":false}`),
		json.RawMessage(`{
			"$type":"app.bsky.actor.defs#savedFeedsPrefV2",
			"items":[{"id":"one","pinned":true,"type":"feed","value":"at://did:plc:one/app.bsky.feed.generator/one"}]
		}`),
		json.RawMessage(`{"$type":"app.bsky.unspecced.defs#unknownPref","keep":true}`),
	})
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}
	return prefs
}

func TestHandler_ContentFilteringPreservesUnknownPrefs(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{prefs: loadedPrefs(t)}
	h := settingscommand.NewHandler(writer)

	got, err := h.HandleContentFiltering(context.Background(), intent.UpdateContentFiltering{
		AdultContentEnabled: true,
		Labels: map[string]string{
			"porn":          "hide",
			"sexual":        "ignore",
			"nudity":        "warn",
			"graphic-media": "hide",
		},
	})
	if err != nil {
		t.Fatalf("HandleContentFiltering() err = %v", err)
	}
	if writer.putCalls != 1 {
		t.Fatalf("putCalls = %d, want 1", writer.putCalls)
	}
	if !got.AdultContentEnabled || got.Labels["sexual"] != bluesky.LabelIgnore {
		t.Fatalf("updated = %#v", got)
	}
	joined := ""
	for _, raw := range writer.put.Raw {
		joined += string(raw)
	}
	if !strings.Contains(joined, "app.bsky.unspecced.defs#unknownPref") {
		t.Fatalf("put Raw = %s, want unknown pref preserved", joined)
	}
	if !strings.Contains(joined, "app.bsky.actor.defs#savedFeedsPrefV2") {
		t.Fatalf("put Raw = %s, want saved feeds preserved", joined)
	}
}

func TestHandler_Threading(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{prefs: loadedPrefs(t)}
	h := settingscommand.NewHandler(writer)

	got, err := h.HandleThreading(context.Background(), intent.UpdateThreading{
		Sort:                    bluesky.ThreadSortOldest,
		PrioritizeFollowedUsers: true,
	})
	if err != nil {
		t.Fatalf("HandleThreading() err = %v", err)
	}
	if got.ThreadView.Sort != bluesky.ThreadSortOldest || !got.ThreadView.PrioritizeFollowedUsers {
		t.Fatalf("ThreadView = %#v", got.ThreadView)
	}
	if writer.putCalls != 1 {
		t.Fatalf("putCalls = %d, want 1", writer.putCalls)
	}
}

func TestHandler_PropagatesGetError(t *testing.T) {
	t.Parallel()

	want := errors.New("pds down")
	h := settingscommand.NewHandler(&stubWriter{getErr: want})
	_, err := h.HandleContentFiltering(context.Background(), intent.UpdateContentFiltering{})
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}

func TestHandler_RequiresWriter(t *testing.T) {
	t.Parallel()

	h := settingscommand.NewHandler(nil)
	if _, err := h.HandleThreading(context.Background(), intent.UpdateThreading{}); err == nil {
		t.Fatal("HandleThreading() err = nil, want error")
	}
}
