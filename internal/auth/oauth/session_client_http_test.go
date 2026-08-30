package oauth_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bluesky-social/indigo/atproto/atclient"
	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/bluesky"
)

const timelineFixture = `{
	"feed": [{
		"post": {
			"uri": "at://did:plc:example/app.bsky.feed.post/abc",
			"author": {"did": "did:plc:example", "handle": "alice.test"},
			"record": {
				"text": "hello timeline",
				"createdAt": "2026-01-15T12:00:00.000Z"
			}
		}
	}],
	"cursor": "page-2"
}`

func TestSessionClient_GetTimeline_UsesAppViewProxy(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/xrpc/app.bsky.feed.getTimeline" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.feed.getTimeline", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "20" {
			t.Fatalf("limit = %q, want 20", got)
		}
		if got := r.Header.Get("Atproto-Proxy"); got != "did:web:api.bsky.app#bsky_appview" {
			t.Fatalf("Atproto-Proxy = %q, want did:web:api.bsky.app#bsky_appview", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(timelineFixture))
	}))
	t.Cleanup(server.Close)

	client := authoauth.NewSessionClient(atclient.NewAPIClient(server.URL))
	resp, err := client.GetTimeline(context.Background(), bluesky.TimelineRequest{Limit: 20})
	if err != nil {
		t.Fatalf("GetTimeline() err = %v", err)
	}
	if resp.Cursor != "page-2" {
		t.Fatalf("Cursor = %q, want page-2", resp.Cursor)
	}
	if len(resp.Feed) != 1 || resp.Feed[0].Post.Record.Text != "hello timeline" {
		t.Fatalf("Feed = %#v, want one hello timeline post", resp.Feed)
	}
}

func TestSessionClient_GetSavedFeeds_UsesPDSWithoutProxy(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/xrpc/app.bsky.actor.getPreferences" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.actor.getPreferences", r.URL.Path)
		}
		if proxy := r.Header.Get("Atproto-Proxy"); proxy != "" {
			t.Fatalf("Atproto-Proxy = %q, want empty for PDS-native getPreferences", proxy)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"preferences": [{
				"$type": "app.bsky.actor.defs#savedFeedsPrefV2",
				"items": [{
					"id": "one",
					"pinned": true,
					"type": "feed",
					"value": "at://did:plc:one/app.bsky.feed.generator/one"
				}]
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := authoauth.NewSessionClient(atclient.NewAPIClient(server.URL))
	feeds, err := client.GetSavedFeeds(context.Background())
	if err != nil {
		t.Fatalf("GetSavedFeeds() err = %v", err)
	}
	if len(feeds) != 1 {
		t.Fatalf("len(feeds) = %d, want 1", len(feeds))
	}
	if feeds[0].URI != "at://did:plc:one/app.bsky.feed.generator/one" {
		t.Fatalf("feeds[0].URI = %q, want pinned feed URI", feeds[0].URI)
	}
}

func TestSessionClient_GetPreferences_UsesPDSWithoutProxy(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/xrpc/app.bsky.actor.getPreferences" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.actor.getPreferences", r.URL.Path)
		}
		if proxy := r.Header.Get("Atproto-Proxy"); proxy != "" {
			t.Fatalf("Atproto-Proxy = %q, want empty", proxy)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"preferences": [
				{"$type":"app.bsky.actor.defs#adultContentPref","enabled":true},
				{"$type":"app.bsky.actor.defs#threadViewPref","sort":"oldest"}
			]
		}`))
	}))
	t.Cleanup(server.Close)

	client := authoauth.NewSessionClient(atclient.NewAPIClient(server.URL))
	prefs, err := client.GetPreferences(context.Background())
	if err != nil {
		t.Fatalf("GetPreferences() err = %v", err)
	}
	if !prefs.AdultContentEnabled {
		t.Fatal("AdultContentEnabled = false, want true")
	}
	if prefs.ThreadView.Sort != bluesky.ThreadSortOldest {
		t.Fatalf("sort = %q, want oldest", prefs.ThreadView.Sort)
	}
}

func TestSessionClient_PutPreferences_UsesPDSWithoutProxy(t *testing.T) {
	t.Parallel()

	var gotPath, gotProxy, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotProxy = r.Header.Get("Atproto-Proxy")
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	prefs, err := bluesky.ParsePreferences(nil)
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}
	updated, err := prefs.WithAdultContent(true)
	if err != nil {
		t.Fatalf("WithAdultContent() err = %v", err)
	}
	client := authoauth.NewSessionClient(atclient.NewAPIClient(server.URL))
	if err := client.PutPreferences(context.Background(), updated); err != nil {
		t.Fatalf("PutPreferences() err = %v", err)
	}
	if gotPath != "/xrpc/app.bsky.actor.putPreferences" {
		t.Fatalf("path = %q, want /xrpc/app.bsky.actor.putPreferences", gotPath)
	}
	if gotProxy != "" {
		t.Fatalf("Atproto-Proxy = %q, want empty", gotProxy)
	}
	if !strings.Contains(gotBody, "app.bsky.actor.defs#adultContentPref") {
		t.Fatalf("body = %q, want adultContentPref", gotBody)
	}
}

func TestSessionClient_GetTimeline_LenientFeedDecode(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "app.bsky.feed.getTimeline") {
			t.Fatalf("path = %q, want getTimeline", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [
				{
					"post": {
						"uri": "at://did:plc:good/app.bsky.feed.post/good1",
						"author": {"did": "did:plc:good", "handle": "good.test"},
						"record": {"text": "kept", "createdAt": "2026-01-15T12:00:00.000Z"}
					}
				},
				{
					"post": {
						"uri": "at://did:plc:badtime/app.bsky.feed.post/bad1",
						"author": {"did": "did:plc:badtime", "handle": "badtime.test"},
						"record": {"text": "also kept", "createdAt": ""}
					}
				},
				{"reason": 42}
			]
		}`))
	}))
	t.Cleanup(server.Close)

	client := authoauth.NewSessionClient(atclient.NewAPIClient(server.URL))
	resp, err := client.GetTimeline(context.Background(), bluesky.TimelineRequest{Limit: 20})
	if err != nil {
		t.Fatalf("GetTimeline() err = %v", err)
	}
	if len(resp.Feed) != 2 {
		t.Fatalf("len(Feed) = %d, want 2 malformed item skipped", len(resp.Feed))
	}
	if resp.Feed[1].Post.Record.Text != "also kept" {
		t.Fatalf("Feed[1].Text = %q, want also kept", resp.Feed[1].Post.Record.Text)
	}
}

const postThreadFixture = `{
	"thread": {
		"$type": "app.bsky.feed.defs#threadViewPost",
		"post": {
			"uri": "at://did:plc:example/app.bsky.feed.post/root",
			"author": {"did": "did:plc:example", "handle": "alice.test"},
			"record": {"text": "hello thread", "createdAt": "2026-01-15T12:00:00.000Z"}
		}
	}
}`

func TestSessionClient_GetPostThread_UsesAppViewProxy(t *testing.T) {
	t.Parallel()

	const postURI = "at://did:plc:example/app.bsky.feed.post/root"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/xrpc/app.bsky.feed.getPostThread" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.feed.getPostThread", r.URL.Path)
		}
		if got := r.URL.Query().Get("uri"); got != postURI {
			t.Fatalf("uri = %q, want %q", got, postURI)
		}
		if got := r.URL.Query().Get("depth"); got != "6" {
			t.Fatalf("depth = %q, want 6", got)
		}
		if got := r.URL.Query().Get("parentHeight"); got != "25" {
			t.Fatalf("parentHeight = %q, want 25", got)
		}
		if got := r.Header.Get("Atproto-Proxy"); got != "did:web:api.bsky.app#bsky_appview" {
			t.Fatalf("Atproto-Proxy = %q, want did:web:api.bsky.app#bsky_appview", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(postThreadFixture))
	}))
	t.Cleanup(server.Close)

	client := authoauth.NewSessionClient(atclient.NewAPIClient(server.URL))
	node, err := client.GetPostThread(context.Background(), postURI, nil)
	if err != nil {
		t.Fatalf("GetPostThread() err = %v", err)
	}
	root, ok := node.(bluesky.ThreadViewPost)
	if !ok {
		t.Fatalf("node type = %T, want ThreadViewPost", node)
	}
	if root.Post.Record.Text != "hello thread" {
		t.Fatalf("text = %q, want hello thread", root.Post.Record.Text)
	}
}

func TestSessionClient_GetProfile_UsesAppViewProxy(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/xrpc/app.bsky.actor.getProfile" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.actor.getProfile", r.URL.Path)
		}
		if got := r.URL.Query().Get("actor"); got != "alice.test" {
			t.Fatalf("actor = %q, want alice.test", got)
		}
		if got := r.Header.Get("Atproto-Proxy"); got != "did:web:api.bsky.app#bsky_appview" {
			t.Fatalf("Atproto-Proxy = %q, want did:web:api.bsky.app#bsky_appview", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"did": "did:plc:alice",
			"handle": "alice.test",
			"displayName": "Alice"
		}`))
	}))
	t.Cleanup(server.Close)

	client := authoauth.NewSessionClient(atclient.NewAPIClient(server.URL))
	profile, err := client.GetProfile(context.Background(), "alice.test")
	if err != nil {
		t.Fatalf("GetProfile() err = %v", err)
	}
	if profile.DID != "did:plc:alice" || profile.Handle != "alice.test" {
		t.Fatalf("profile = %#v, want alice", profile)
	}
}
