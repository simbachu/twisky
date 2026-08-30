package bluesky_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
)

func samplePrefs() []json.RawMessage {
	return []json.RawMessage{
		json.RawMessage(`{"$type":"app.bsky.actor.defs#adultContentPref","enabled":false}`),
		json.RawMessage(`{"$type":"app.bsky.actor.defs#contentLabelPref","label":"porn","visibility":"show"}`),
		json.RawMessage(`{"$type":"app.bsky.actor.defs#contentLabelPref","labelerDid":"did:plc:labeler","label":"porn","visibility":"hide"}`),
		json.RawMessage(`{
			"$type":"app.bsky.actor.defs#savedFeedsPrefV2",
			"items":[{"id":"one","pinned":true,"type":"feed","value":"at://did:plc:one/app.bsky.feed.generator/one"}]
		}`),
		json.RawMessage(`{"$type":"app.bsky.actor.defs#threadViewPref","sort":"newest","prioritizeFollowedUsers":true}`),
		json.RawMessage(`{"$type":"app.bsky.unspecced.defs#unknownPref","keep":true}`),
	}
}

func TestParsePreferences_ShowMapsToIgnore(t *testing.T) {
	t.Parallel()

	prefs, err := bluesky.ParsePreferences(samplePrefs())
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}
	if prefs.Labels["porn"] != bluesky.LabelIgnore {
		t.Fatalf("porn = %q, want ignore (from show)", prefs.Labels["porn"])
	}
	if prefs.Labels["sexual"] != bluesky.LabelWarn {
		t.Fatalf("sexual = %q, want default warn", prefs.Labels["sexual"])
	}
}

func TestParsePreferences_ParsesAdultContentThreadAndFeeds(t *testing.T) {
	t.Parallel()

	prefs, err := bluesky.ParsePreferences(samplePrefs())
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}
	if prefs.AdultContentEnabled {
		t.Fatal("AdultContentEnabled = true, want false")
	}
	if prefs.ThreadView.Sort != bluesky.ThreadSortNewest || !prefs.ThreadView.PrioritizeFollowedUsers {
		t.Fatalf("ThreadView = %#v", prefs.ThreadView)
	}
	if len(prefs.SavedFeeds) != 1 || prefs.SavedFeeds[0].ID != "one" {
		t.Fatalf("SavedFeeds = %#v", prefs.SavedFeeds)
	}
	if len(prefs.Raw) != 6 {
		t.Fatalf("len(Raw) = %d, want 6", len(prefs.Raw))
	}
}

func TestParsePreferences_SavedFeedsLegacy(t *testing.T) {
	t.Parallel()

	prefs, err := bluesky.ParsePreferences([]json.RawMessage{json.RawMessage(`{
		"$type":"app.bsky.actor.defs#savedFeedsPref",
		"pinned":["at://did:plc:one/app.bsky.feed.generator/one"],
		"saved":[]
	}`)})
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}
	if len(prefs.SavedFeeds) != 1 || !prefs.SavedFeeds[0].Pinned || prefs.SavedFeeds[0].Type != "feed" {
		t.Fatalf("SavedFeeds = %#v, want pinned legacy feed", prefs.SavedFeeds)
	}
}

func TestWithAdultContent_PreservesUnknownAndSavedFeeds(t *testing.T) {
	t.Parallel()

	prefs, err := bluesky.ParsePreferences(samplePrefs())
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}

	updated, err := prefs.WithAdultContent(true)
	if err != nil {
		t.Fatalf("WithAdultContent() err = %v", err)
	}
	if !updated.AdultContentEnabled {
		t.Fatal("AdultContentEnabled = false, want true")
	}
	if len(updated.SavedFeeds) != 1 || updated.SavedFeeds[0].ID != "one" {
		t.Fatalf("SavedFeeds dropped: %#v", updated.SavedFeeds)
	}
	joined := string(bytesJoin(updated.Raw))
	if !strings.Contains(joined, "app.bsky.unspecced.defs#unknownPref") {
		t.Fatalf("Raw = %s, want unknown pref preserved", joined)
	}
	if !strings.Contains(joined, `"enabled":true`) {
		t.Fatalf("Raw = %s, want enabled true", joined)
	}
	if strings.Count(joined, "app.bsky.actor.defs#adultContentPref") != 1 {
		t.Fatalf("Raw = %s, want one adultContentPref", joined)
	}
}

func TestWithContentLabels_PreservesLabelerSpecificPrefs(t *testing.T) {
	t.Parallel()

	prefs, err := bluesky.ParsePreferences(samplePrefs())
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}

	updated, err := prefs.WithContentLabels(map[string]bluesky.LabelVisibility{
		"porn":          bluesky.LabelHide,
		"sexual":        bluesky.LabelIgnore,
		"nudity":        bluesky.LabelWarn,
		"graphic-media": bluesky.LabelHide,
	})
	if err != nil {
		t.Fatalf("WithContentLabels() err = %v", err)
	}
	if updated.Labels["porn"] != bluesky.LabelHide || updated.Labels["sexual"] != bluesky.LabelIgnore {
		t.Fatalf("Labels = %#v", updated.Labels)
	}
	joined := string(bytesJoin(updated.Raw))
	if !strings.Contains(joined, `"labelerDid":"did:plc:labeler"`) {
		t.Fatalf("Raw = %s, want labeler-specific porn pref kept", joined)
	}
	if !strings.Contains(joined, "app.bsky.actor.defs#savedFeedsPrefV2") {
		t.Fatalf("Raw = %s, want saved feeds kept", joined)
	}
}

func TestWithThreadView_ReplacesOnlyThreadPref(t *testing.T) {
	t.Parallel()

	prefs, err := bluesky.ParsePreferences(samplePrefs())
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}

	updated, err := prefs.WithThreadView(bluesky.ThreadViewPref{
		Sort:                    bluesky.ThreadSortOldest,
		PrioritizeFollowedUsers: false,
	})
	if err != nil {
		t.Fatalf("WithThreadView() err = %v", err)
	}
	if updated.ThreadView.Sort != bluesky.ThreadSortOldest || updated.ThreadView.PrioritizeFollowedUsers {
		t.Fatalf("ThreadView = %#v", updated.ThreadView)
	}
	if updated.AdultContentEnabled {
		t.Fatal("adult content should be unchanged")
	}
	joined := string(bytesJoin(updated.Raw))
	if strings.Count(joined, "app.bsky.actor.defs#threadViewPref") != 1 {
		t.Fatalf("Raw = %s, want one threadViewPref", joined)
	}
}

func bytesJoin(raw []json.RawMessage) []byte {
	var b []byte
	for _, item := range raw {
		b = append(b, item...)
	}
	return b
}
