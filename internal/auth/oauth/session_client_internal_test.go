package oauth

import (
	"encoding/json"
	"testing"
)

func TestParseSavedFeedsPrefV2(t *testing.T) {
	t.Parallel()

	preferences := []json.RawMessage{
		json.RawMessage(`{"$type":"app.bsky.actor.defs#adultContentPref","enabled":false}`),
		json.RawMessage(`{
			"$type":"app.bsky.actor.defs#savedFeedsPrefV2",
			"items":[
				{"id":"one","pinned":true,"type":"feed","value":"at://did:plc:one/app.bsky.feed.generator/one"},
				{"id":"two","pinned":false,"type":"feed","value":"at://did:plc:two/app.bsky.feed.generator/two"}
			]
		}`),
	}

	feeds, err := parseSavedFeeds(preferences)

	if err != nil {
		t.Fatalf("parseSavedFeeds() err = %v", err)
	}
	if len(feeds) != 2 {
		t.Fatalf("len(feeds) = %d, want 2", len(feeds))
	}
	if feeds[0].ID != "one" || feeds[0].URI != "at://did:plc:one/app.bsky.feed.generator/one" {
		t.Fatalf("feeds[0] = %#v, want parsed V2 item", feeds[0])
	}
}

func TestParseSavedFeedsPrefLegacy(t *testing.T) {
	t.Parallel()

	preferences := []json.RawMessage{json.RawMessage(`{
		"$type":"app.bsky.actor.defs#savedFeedsPref",
		"pinned":["at://did:plc:one/app.bsky.feed.generator/one"],
		"saved":[]
	}`)}

	feeds, err := parseSavedFeeds(preferences)

	if err != nil {
		t.Fatalf("parseSavedFeeds() err = %v", err)
	}
	if len(feeds) != 1 {
		t.Fatalf("len(feeds) = %d, want 1", len(feeds))
	}
	if !feeds[0].Pinned || feeds[0].Type != "feed" {
		t.Fatalf("feeds[0] = %#v, want pinned feed", feeds[0])
	}
}
