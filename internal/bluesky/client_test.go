package bluesky_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
)

func TestClient_GetAuthorFeed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xrpc/app.bsky.feed.getAuthorFeed" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.feed.getAuthorFeed", r.URL.Path)
		}
		if got := r.URL.Query().Get("actor"); got != "bsky.app" {
			t.Fatalf("actor = %q, want bsky.app", got)
		}
		if got := r.URL.Query().Get("filter"); got != bluesky.FilterPostsWithMedia {
			t.Fatalf("filter = %q, want %s", got, bluesky.FilterPostsWithMedia)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/abc",
					"author": {
						"handle": "bsky.app",
						"displayName": "Bluesky",
						"avatar": "https://example.com/avatar.jpg"
					},
					"record": {
						"text": "hello with image",
						"createdAt": "2026-01-15T12:00:00.000Z"
					},
					"embed": {
						"$type": "app.bsky.embed.images#view",
						"images": [{
							"thumb": "https://example.com/thumb.jpg",
							"fullsize": "https://example.com/full.jpg",
							"alt": "a photo",
							"aspectRatio": {"width": 4000, "height": 3000}
						}]
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{
		Actor:  "bsky.app",
		Filter: bluesky.FilterPostsWithMedia,
	})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}
	if len(items.Feed) != 1 {
		t.Fatalf("len(items.Feed) = %d, want 1", len(items.Feed))
	}

	post := items.Feed[0].Post
	if post.Record.Text != "hello with image" {
		t.Fatalf("post.Record.Text = %q, want hello with image", post.Record.Text)
	}
	if post.Embed == nil || len(post.Embed.Images) != 1 {
		t.Fatalf("post.Embed.Images = %#v, want one image", post.Embed)
	}
	if post.Embed.Images[0].Thumb != "https://example.com/thumb.jpg" {
		t.Fatalf("thumb = %q, want https://example.com/thumb.jpg", post.Embed.Images[0].Thumb)
	}
	if post.Embed.Images[0].Fullsize != "https://example.com/full.jpg" {
		t.Fatalf("fullsize = %q, want https://example.com/full.jpg", post.Embed.Images[0].Fullsize)
	}
	if post.Embed.Images[0].Alt != "a photo" {
		t.Fatalf("alt = %q, want a photo", post.Embed.Images[0].Alt)
	}
	if post.Embed.Images[0].AspectRatio == nil || post.Embed.Images[0].AspectRatio.Width != 4000 {
		t.Fatalf("aspectRatio = %#v, want width 4000", post.Embed.Images[0].AspectRatio)
	}
}

func TestClient_GetAuthorFeed_GalleryEmbed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/gallery",
					"author": {
						"handle": "bsky.app",
						"displayName": "Bluesky"
					},
					"record": {
						"text": "gallery post",
						"createdAt": "2026-01-15T12:00:00.000Z"
					},
					"embed": {
						"$type": "app.bsky.embed.gallery#view",
						"items": [
							{
								"thumbnail": "https://example.com/thumb1.jpg",
								"fullsize": "https://example.com/full1.jpg",
								"alt": "",
								"aspectRatio": {"width": 1000, "height": 800}
							},
							{
								"thumbnail": "https://example.com/thumb2.jpg",
								"fullsize": "https://example.com/full2.jpg",
								"alt": "second",
								"aspectRatio": {"width": 1200, "height": 900}
							}
						]
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{
		Actor:  "bsky.app",
		Filter: bluesky.FilterPostsWithMedia,
	})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	post := items.Feed[0].Post
	images := post.Embed.MediaImages()
	if len(images) != 2 {
		t.Fatalf("len(MediaImages()) = %d, want 2", len(images))
	}
	if got := images[0].ThumbURL(); got != "https://example.com/thumb1.jpg" {
		t.Fatalf("images[0].ThumbURL() = %q, want https://example.com/thumb1.jpg", got)
	}
	if images[1].Fullsize != "https://example.com/full2.jpg" {
		t.Fatalf("images[1].Fullsize = %q, want https://example.com/full2.jpg", images[1].Fullsize)
	}
	if images[1].Alt != "second" {
		t.Fatalf("images[1].Alt = %q, want second", images[1].Alt)
	}
}

func TestClient_GetAuthorFeed_PassesLimitAndCursor(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "20" {
			t.Fatalf("limit = %q, want 20", got)
		}
		if got := r.URL.Query().Get("cursor"); got != "next-page" {
			t.Fatalf("cursor = %q, want next-page", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"feed":[],"cursor":"another-page"}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	resp, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{
		Actor:  "bsky.app",
		Filter: bluesky.FilterPostsNoReplies,
		Limit:  20,
		Cursor: "next-page",
	})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}
	if resp.Cursor != "another-page" {
		t.Fatalf("Cursor = %q, want another-page", resp.Cursor)
	}
}

func TestClient_GetAuthorFeed_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	_, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{
		Actor:  "missing.example",
		Filter: bluesky.FilterPostsNoReplies,
	})
	if err == nil {
		t.Fatal("GetAuthorFeed() err = nil, want ErrNotFound")
	}
	if !errors.Is(err, bluesky.ErrNotFound) {
		t.Fatalf("GetAuthorFeed() err = %v, want ErrNotFound", err)
	}
}

func TestClient_GetAuthorFeed_ParsesCreatedAt(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/abc",
					"author": {"handle": "bsky.app"},
					"record": {
						"text": "timed post",
						"createdAt": "2026-01-15T12:00:00.000Z"
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{
		Actor:  "bsky.app",
		Filter: bluesky.FilterPostsNoReplies,
	})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	want := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	if !items.Feed[0].Post.Record.CreatedAt.Equal(want) {
		t.Fatalf("CreatedAt = %v, want %v", items.Feed[0].Post.Record.CreatedAt, want)
	}
}

func TestClient_SearchPosts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xrpc/app.bsky.feed.searchPosts" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.feed.searchPosts", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != "#golang" {
			t.Fatalf("q = %q, want #golang", got)
		}
		if got := r.URL.Query().Get("sort"); got != "latest" {
			t.Fatalf("sort = %q, want latest", got)
		}
		if r.URL.Query().Get("tag") != "" {
			t.Fatalf("tag param should not be set")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"posts": [{
				"uri": "at://did:plc:example/app.bsky.feed.post/abc",
				"author": {
					"handle": "dev.example",
					"displayName": "Developer",
					"avatar": "https://example.com/avatar.jpg"
				},
				"record": {
					"text": "hello #golang",
					"createdAt": "2026-01-15T12:00:00.000Z",
					"facets": [{
						"index": {"byteStart": 6, "byteEnd": 13},
						"features": [{
							"$type": "app.bsky.richtext.facet#tag",
							"tag": "golang"
						}]
					}, {
						"index": {"byteStart": 0, "byteEnd": 9},
						"features": [{
							"$type": "app.bsky.richtext.facet#mention",
							"did": "did:plc:example"
						}]
					}, {
						"index": {"byteStart": 20, "byteEnd": 39},
						"features": [{
							"$type": "app.bsky.richtext.facet#link",
							"uri": "https://example.com"
						}]
					}]
				}
			}],
			"cursor": "next-page"
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	resp, err := client.SearchPosts(context.Background(), bluesky.SearchPostsRequest{
		Tag: "golang",
	})
	if err != nil {
		t.Fatalf("SearchPosts() err = %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Fatalf("len(resp.Posts) = %d, want 1", len(resp.Posts))
	}
	if resp.Posts[0].Record.Text != "hello #golang" {
		t.Fatalf("post text = %q, want hello #golang", resp.Posts[0].Record.Text)
	}
	if len(resp.Posts[0].Record.Facets) != 3 {
		t.Fatalf("len(facets) = %d, want 3", len(resp.Posts[0].Record.Facets))
	}
	if resp.Posts[0].Record.Facets[0].Features[0].Tag != "golang" {
		t.Fatalf("facet tag = %q, want golang", resp.Posts[0].Record.Facets[0].Features[0].Tag)
	}
	if resp.Posts[0].Record.Facets[1].Features[0].DID != "did:plc:example" {
		t.Fatalf("mention did = %q, want did:plc:example", resp.Posts[0].Record.Facets[1].Features[0].DID)
	}
	if resp.Posts[0].Record.Facets[2].Features[0].URI != "https://example.com" {
		t.Fatalf("link uri = %q, want https://example.com", resp.Posts[0].Record.Facets[2].Features[0].URI)
	}
	if resp.Cursor != "next-page" {
		t.Fatalf("Cursor = %q, want next-page", resp.Cursor)
	}
}

func TestClient_SearchPosts_PassesTagAndLimit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "#golang" {
			t.Fatalf("q = %q, want #golang", got)
		}
		if got := r.URL.Query().Get("limit"); got != "20" {
			t.Fatalf("limit = %q, want 20", got)
		}
		if got := r.URL.Query().Get("cursor"); got != "next-page" {
			t.Fatalf("cursor = %q, want next-page", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"posts":[],"cursor":"another-page"}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	resp, err := client.SearchPosts(context.Background(), bluesky.SearchPostsRequest{
		Tag:    "golang",
		Limit:  20,
		Cursor: "next-page",
	})
	if err != nil {
		t.Fatalf("SearchPosts() err = %v", err)
	}
	if resp.Cursor != "another-page" {
		t.Fatalf("Cursor = %q, want another-page", resp.Cursor)
	}
}

func TestClient_GetProfiles(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xrpc/app.bsky.actor.getProfiles" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.actor.getProfiles", r.URL.Path)
		}
		actors := r.URL.Query()["actors"]
		if len(actors) != 2 {
			t.Fatalf("len(actors) = %d, want 2", len(actors))
		}
		if actors[0] != "did:plc:one" || actors[1] != "did:plc:two" {
			t.Fatalf("actors = %v, want [did:plc:one did:plc:two]", actors)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"profiles": [
				{"did": "did:plc:one", "handle": "one.example"},
				{"did": "did:plc:two", "handle": "two.example"}
			]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	profiles, err := client.GetProfiles(context.Background(), []string{"did:plc:one", "did:plc:two"})
	if err != nil {
		t.Fatalf("GetProfiles() err = %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("len(profiles) = %d, want 2", len(profiles))
	}
	if profiles[0].Handle != "one.example" {
		t.Fatalf("profiles[0].Handle = %q, want one.example", profiles[0].Handle)
	}
}

func TestClient_GetPostThread(t *testing.T) {
	t.Parallel()

	const postURI = "at://did:plc:example/app.bsky.feed.post/root"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xrpc/app.bsky.feed.getPostThread" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.feed.getPostThread", r.URL.Path)
		}
		if got := r.URL.Query().Get("uri"); got != postURI {
			t.Fatalf("uri = %q, want %s", got, postURI)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"thread": {
				"$type": "app.bsky.feed.defs#threadViewPost",
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/root",
					"author": {"handle": "bsky.app", "displayName": "Bluesky"},
					"record": {"text": "root post", "createdAt": "2026-01-15T12:00:00.000Z"}
				},
				"parent": {
					"$type": "app.bsky.feed.defs#threadViewPost",
					"post": {
						"uri": "at://did:plc:example/app.bsky.feed.post/parent",
						"author": {"handle": "bsky.app", "displayName": "Bluesky"},
						"record": {"text": "parent post", "createdAt": "2026-01-15T11:00:00.000Z"}
					}
				},
				"replies": [{
					"$type": "app.bsky.feed.defs#threadViewPost",
					"post": {
						"uri": "at://did:plc:example/app.bsky.feed.post/reply1",
						"author": {"handle": "dev.example", "displayName": "Dev"},
						"record": {"text": "reply one", "createdAt": "2026-01-15T13:00:00.000Z"}
					},
					"replies": [{
						"$type": "app.bsky.feed.defs#threadViewPost",
						"post": {
							"uri": "at://did:plc:example/app.bsky.feed.post/reply2",
							"author": {"handle": "dev.example", "displayName": "Dev"},
							"record": {"text": "nested reply", "createdAt": "2026-01-15T14:00:00.000Z"}
						}
					}]
				}, {
					"$type": "app.bsky.feed.defs#notFoundPost",
					"uri": "at://did:plc:example/app.bsky.feed.post/missing",
					"notFound": true
				}]
			}
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	node, err := client.GetPostThread(context.Background(), postURI)
	if err != nil {
		t.Fatalf("GetPostThread() err = %v", err)
	}

	root, ok := node.(bluesky.ThreadViewPost)
	if !ok {
		t.Fatalf("node type = %T, want ThreadViewPost", node)
	}
	if root.Post.Record.Text != "root post" {
		t.Fatalf("root text = %q, want root post", root.Post.Record.Text)
	}

	parent, ok := root.Parent.(bluesky.ThreadViewPost)
	if !ok {
		t.Fatalf("parent type = %T, want ThreadViewPost", root.Parent)
	}
	if parent.Post.Record.Text != "parent post" {
		t.Fatalf("parent text = %q, want parent post", parent.Post.Record.Text)
	}

	if len(root.Replies) != 2 {
		t.Fatalf("len(root.Replies) = %d, want 2", len(root.Replies))
	}

	reply, ok := root.Replies[0].(bluesky.ThreadViewPost)
	if !ok {
		t.Fatalf("reply type = %T, want ThreadViewPost", root.Replies[0])
	}
	if reply.Post.Record.Text != "reply one" {
		t.Fatalf("reply text = %q, want reply one", reply.Post.Record.Text)
	}
	if len(reply.Replies) != 1 {
		t.Fatalf("len(reply.Replies) = %d, want 1", len(reply.Replies))
	}

	nested, ok := reply.Replies[0].(bluesky.ThreadViewPost)
	if !ok {
		t.Fatalf("nested reply type = %T, want ThreadViewPost", reply.Replies[0])
	}
	if nested.Post.Record.Text != "nested reply" {
		t.Fatalf("nested reply text = %q, want nested reply", nested.Post.Record.Text)
	}

	if _, ok := root.Replies[1].(bluesky.NotFoundPost); !ok {
		t.Fatalf("second reply type = %T, want NotFoundPost", root.Replies[1])
	}
}

func TestClient_GetPostThread_NotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	_, err := client.GetPostThread(context.Background(), "at://did:plc:example/app.bsky.feed.post/missing")
	if err == nil {
		t.Fatal("GetPostThread() err = nil, want ErrNotFound")
	}
	if !errors.Is(err, bluesky.ErrNotFound) {
		t.Fatalf("GetPostThread() err = %v, want ErrNotFound", err)
	}
}
