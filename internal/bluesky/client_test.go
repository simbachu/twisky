package bluesky_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestClient_GetProfile_ParsesPinnedPost(t *testing.T) {
	t.Parallel()

	const pinnedURI = "at://did:plc:example/app.bsky.feed.post/pinned123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xrpc/app.bsky.actor.getProfile" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.actor.getProfile", r.URL.Path)
		}
		if got := r.URL.Query().Get("actor"); got != "bsky.app" {
			t.Fatalf("actor = %q, want bsky.app", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"did": "did:plc:example",
			"handle": "bsky.app",
			"displayName": "Bluesky",
			"pinnedPost": {
				"uri": "` + pinnedURI + `",
				"cid": "bafyreicid"
			}
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	profile, err := client.GetProfile(context.Background(), "bsky.app")
	if err != nil {
		t.Fatalf("GetProfile() err = %v", err)
	}
	if profile.PinnedPost == nil {
		t.Fatal("PinnedPost = nil, want strong ref")
	}
	if profile.PinnedPost.URI != pinnedURI {
		t.Fatalf("PinnedPost.URI = %q, want %s", profile.PinnedPost.URI, pinnedURI)
	}
	if profile.PinnedPost.CID != "bafyreicid" {
		t.Fatalf("PinnedPost.CID = %q, want bafyreicid", profile.PinnedPost.CID)
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

func TestClient_GetAuthorFeed_RecordEmbed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/qrt",
					"author": {"handle": "dev.example", "displayName": "Dev"},
					"record": {
						"text": "my take",
						"createdAt": "2026-01-15T12:00:00.000Z"
					},
					"embed": {
						"$type": "app.bsky.embed.record#view",
						"record": {
							"$type": "app.bsky.embed.record#viewRecord",
							"uri": "at://did:plc:quoted/app.bsky.feed.post/original",
							"author": {"handle": "quoted.example", "displayName": "Quoted"},
							"value": {
								"text": "original post",
								"createdAt": "2026-01-14T12:00:00.000Z"
							}
						}
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{
		Actor: "dev.example",
	})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	quoted := items.Feed[0].Post.Embed.QuotedPost()
	if quoted == nil {
		t.Fatal("QuotedPost() = nil, want quoted post")
	}
	if quoted.URI != "at://did:plc:quoted/app.bsky.feed.post/original" {
		t.Fatalf("quoted.URI = %q, want original URI", quoted.URI)
	}
	if quoted.Author.Handle != "quoted.example" {
		t.Fatalf("quoted.Author.Handle = %q, want quoted.example", quoted.Author.Handle)
	}
	if quoted.Record.Text != "original post" {
		t.Fatalf("quoted.Record.Text = %q, want original post", quoted.Record.Text)
	}
}

func TestClient_GetAuthorFeed_ExternalEmbed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/link",
					"author": {"handle": "dev.example"},
					"record": {
						"text": "check this out https://example.com",
						"createdAt": "2026-01-15T12:00:00.000Z"
					},
					"embed": {
						"$type": "app.bsky.embed.external#view",
						"external": {
							"uri": "https://example.com",
							"title": "Example Site",
							"description": "An example website",
							"thumb": "https://example.com/thumb.jpg"
						}
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{
		Actor: "dev.example",
	})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	external := items.Feed[0].Post.Embed.ExternalLink()
	if external == nil {
		t.Fatal("ExternalLink() = nil, want external card")
	}
	if external.URI != "https://example.com" {
		t.Fatalf("external.URI = %q, want https://example.com", external.URI)
	}
	if external.Title != "Example Site" {
		t.Fatalf("external.Title = %q, want Example Site", external.Title)
	}
	if external.Description != "An example website" {
		t.Fatalf("external.Description = %q, want An example website", external.Description)
	}
	if external.Thumb != "https://example.com/thumb.jpg" {
		t.Fatalf("external.Thumb = %q, want thumb URL", external.Thumb)
	}
}

func TestClient_GetAuthorFeed_NestedExternalInQuotedPost(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/qrt-link",
					"author": {"handle": "dev.example"},
					"record": {
						"text": "sharing news",
						"createdAt": "2026-01-15T12:00:00.000Z"
					},
					"embed": {
						"$type": "app.bsky.embed.record#view",
						"record": {
							"$type": "app.bsky.embed.record#viewRecord",
							"uri": "at://did:plc:quoted/app.bsky.feed.post/with-link",
							"author": {"handle": "quoted.example"},
							"value": {
								"text": "read more https://example.com/article",
								"createdAt": "2026-01-14T12:00:00.000Z"
							},
							"embeds": [{
								"$type": "app.bsky.embed.external#view",
								"external": {
									"uri": "https://example.com/article",
									"title": "Article Title",
									"description": "Article summary"
								}
							}]
						}
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{
		Actor: "dev.example",
	})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	quoted := items.Feed[0].Post.Embed.QuotedPost()
	if quoted == nil {
		t.Fatal("QuotedPost() = nil, want quoted post")
	}
	if quoted.Embed == nil {
		t.Fatal("quoted.Embed = nil, want nested external embed")
	}
	external := quoted.Embed.ExternalLink()
	if external == nil {
		t.Fatal("ExternalLink() = nil, want nested external card")
	}
	if external.URI != "https://example.com/article" {
		t.Fatalf("external.URI = %q, want article URL", external.URI)
	}
	if external.Title != "Article Title" {
		t.Fatalf("external.Title = %q, want Article Title", external.Title)
	}
}

func TestClient_GetAuthorFeed_RecordWithMediaEmbed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/qrt-media",
					"author": {"handle": "dev.example"},
					"record": {"text": "quote with media", "createdAt": "2026-01-15T12:00:00.000Z"},
					"embed": {
						"$type": "app.bsky.embed.recordWithMedia#view",
						"record": {
							"$type": "app.bsky.embed.record#viewRecord",
							"uri": "at://did:plc:quoted/app.bsky.feed.post/with-image",
							"author": {"handle": "quoted.example"},
							"value": {"text": "has image", "createdAt": "2026-01-14T12:00:00.000Z"},
							"embeds": [{
								"$type": "app.bsky.embed.images#view",
								"images": [{
									"thumb": "https://example.com/thumb.jpg",
									"fullsize": "https://example.com/full.jpg",
									"alt": "quoted photo"
								}]
							}]
						},
						"media": {
							"$type": "app.bsky.embed.images#view",
							"images": [{
								"thumb": "https://example.com/my-thumb.jpg",
								"fullsize": "https://example.com/my-full.jpg",
								"alt": "my photo"
							}]
						}
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{Actor: "dev.example"})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	post := items.Feed[0].Post
	media := post.Embed.MediaImages()
	if len(media) != 1 || media[0].Alt != "my photo" {
		t.Fatalf("post media = %#v, want my photo", media)
	}

	quoted := post.Embed.QuotedPost()
	if quoted == nil {
		t.Fatal("QuotedPost() = nil, want quoted post")
	}
	if quoted.Record.Text != "has image" {
		t.Fatalf("quoted text = %q, want has image", quoted.Record.Text)
	}
	quotedImages := quoted.Embed.MediaImages()
	if len(quotedImages) != 1 || quotedImages[0].Alt != "quoted photo" {
		t.Fatalf("quoted images = %#v, want quoted photo", quotedImages)
	}
}

func TestClient_GetAuthorFeed_VideoEmbed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/video",
					"author": {"handle": "dev.example"},
					"record": {"text": "check this clip", "createdAt": "2026-01-15T12:00:00.000Z"},
					"embed": {
						"$type": "app.bsky.embed.video#view",
						"cid": "bafyvideocid",
						"playlist": "https://video.example.com/playlist.m3u8",
						"thumbnail": "https://video.example.com/thumb.jpg",
						"alt": "a short clip",
						"aspectRatio": {"width": 1920, "height": 1080},
						"presentation": "default"
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{Actor: "dev.example"})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	videos := items.Feed[0].Post.Embed.MediaVideos()
	if len(videos) != 1 {
		t.Fatalf("len(videos) = %d, want 1", len(videos))
	}
	video := videos[0]
	if video.Playlist != "https://video.example.com/playlist.m3u8" {
		t.Fatalf("video.Playlist = %q, want playlist URL", video.Playlist)
	}
	if video.Thumbnail != "https://video.example.com/thumb.jpg" {
		t.Fatalf("video.Thumbnail = %q, want thumbnail URL", video.Thumbnail)
	}
	if video.Alt != "a short clip" {
		t.Fatalf("video.Alt = %q, want a short clip", video.Alt)
	}
	if video.AspectRatio == nil || video.AspectRatio.Width != 1920 || video.AspectRatio.Height != 1080 {
		t.Fatalf("video.AspectRatio = %#v, want 1920x1080", video.AspectRatio)
	}
	if video.Presentation != "default" {
		t.Fatalf("video.Presentation = %q, want default", video.Presentation)
	}
}

func TestClient_GetAuthorFeed_RecordWithMediaVideoEmbed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/qrt-video",
					"author": {"handle": "dev.example"},
					"record": {"text": "quote with video", "createdAt": "2026-01-15T12:00:00.000Z"},
					"embed": {
						"$type": "app.bsky.embed.recordWithMedia#view",
						"record": {
							"$type": "app.bsky.embed.record#viewRecord",
							"uri": "at://did:plc:quoted/app.bsky.feed.post/with-video",
							"author": {"handle": "quoted.example"},
							"value": {"text": "has video", "createdAt": "2026-01-14T12:00:00.000Z"},
							"embeds": [{
								"$type": "app.bsky.embed.video#view",
								"cid": "bafyquotedvideo",
								"playlist": "https://video.example.com/quoted.m3u8",
								"thumbnail": "https://video.example.com/quoted-thumb.jpg",
								"alt": "quoted clip"
							}]
						},
						"media": {
							"$type": "app.bsky.embed.video#view",
							"cid": "bafymyvideo",
							"playlist": "https://video.example.com/my.m3u8",
							"thumbnail": "https://video.example.com/my-thumb.jpg",
							"alt": "my clip",
							"presentation": "gif"
						}
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{Actor: "dev.example"})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	post := items.Feed[0].Post
	videos := post.Embed.MediaVideos()
	if len(videos) != 1 || videos[0].Alt != "my clip" {
		t.Fatalf("post videos = %#v, want my clip", videos)
	}
	if !videos[0].IsGIF() {
		t.Fatal("video.IsGIF() = false, want true for gif presentation")
	}

	quoted := post.Embed.QuotedPost()
	if quoted == nil {
		t.Fatal("QuotedPost() = nil, want quoted post")
	}
	if quoted.Record.Text != "has video" {
		t.Fatalf("quoted text = %q, want has video", quoted.Record.Text)
	}
	quotedVideos := quoted.Embed.MediaVideos()
	if len(quotedVideos) != 1 || quotedVideos[0].Alt != "quoted clip" {
		t.Fatalf("quoted videos = %#v, want quoted clip", quotedVideos)
	}
}

func TestClient_GetAuthorFeed_ReplyRefOnRecord(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/reply",
					"author": {"handle": "dev.example"},
					"record": {
						"text": "a reply",
						"createdAt": "2026-01-15T12:00:00.000Z",
						"reply": {
							"root": {"uri": "at://did:plc:example/app.bsky.feed.post/root", "cid": "bafyroot"},
							"parent": {"uri": "at://did:plc:example/app.bsky.feed.post/parent", "cid": "bafyparent"}
						}
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{Actor: "dev.example"})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	got := items.Feed[0].Post.ReplyParentURI()
	want := "at://did:plc:example/app.bsky.feed.post/parent"
	if got != want {
		t.Fatalf("ReplyParentURI() = %q, want %q", got, want)
	}
}

func TestClient_GetAuthorFeed_FeedItemReplyParent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/reply",
					"author": {"handle": "dev.example"},
					"record": {"text": "a reply", "createdAt": "2026-01-15T12:00:00.000Z"}
				},
				"reply": {
					"root": {
						"uri": "at://did:plc:example/app.bsky.feed.post/root",
						"author": {"handle": "dev.example"},
						"record": {"text": "root", "createdAt": "2026-01-15T11:00:00.000Z"}
					},
					"parent": {
						"uri": "at://did:plc:example/app.bsky.feed.post/parent",
						"author": {"handle": "other.example"},
						"record": {"text": "parent post", "createdAt": "2026-01-15T11:30:00.000Z"}
					}
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{Actor: "dev.example"})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	if items.Feed[0].Reply == nil || items.Feed[0].Reply.Parent == nil {
		t.Fatal("Reply.Parent = nil, want hydrated parent post")
	}
	if items.Feed[0].Reply.Parent.Record.Text != "parent post" {
		t.Fatalf("parent text = %q, want parent post", items.Feed[0].Reply.Parent.Record.Text)
	}
}

func TestClient_GetAuthorFeed_FeedItemRepostReason(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"feed": [{
				"post": {
					"uri": "at://did:plc:example/app.bsky.feed.post/original",
					"author": {"handle": "original.example", "displayName": "Original"},
					"record": {"text": "original post", "createdAt": "2026-01-15T12:00:00.000Z"}
				},
				"reason": {
					"$type": "app.bsky.feed.defs#reasonRepost",
					"by": {"handle": "reposter.example", "displayName": "Reposter"},
					"indexedAt": "2026-01-15T13:00:00.000Z"
				}
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	items, err := client.GetAuthorFeed(context.Background(), bluesky.AuthorFeedRequest{Actor: "reposter.example"})
	if err != nil {
		t.Fatalf("GetAuthorFeed() err = %v", err)
	}

	if items.Feed[0].Reason == nil {
		t.Fatal("Reason = nil, want repost reason")
	}
	if items.Feed[0].Reason.RepostedBy.Handle != "reposter.example" {
		t.Fatalf("RepostedBy.Handle = %q, want reposter.example", items.Feed[0].Reason.RepostedBy.Handle)
	}
}

func TestClient_GetPosts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/xrpc/app.bsky.feed.getPosts" {
			t.Fatalf("path = %q, want /xrpc/app.bsky.feed.getPosts", r.URL.Path)
		}
		uris := r.URL.Query()["uris"]
		if len(uris) != 2 {
			t.Fatalf("len(uris) = %d, want 2", len(uris))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"posts": [
				{
					"uri": "at://did:plc:example/app.bsky.feed.post/parent",
					"author": {"handle": "dev.example"},
					"record": {"text": "parent post", "createdAt": "2026-01-15T11:00:00.000Z"}
				},
				{
					"uri": "at://did:plc:example/app.bsky.feed.post/other",
					"author": {"handle": "other.example"},
					"record": {"text": "other post", "createdAt": "2026-01-15T12:00:00.000Z"}
				}
			]
		}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	posts, err := client.GetPosts(context.Background(), []string{
		"at://did:plc:example/app.bsky.feed.post/parent",
		"at://did:plc:example/app.bsky.feed.post/other",
	})
	if err != nil {
		t.Fatalf("GetPosts() err = %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("len(posts) = %d, want 2", len(posts))
	}
	if posts[0].Record.Text != "parent post" {
		t.Fatalf("posts[0].text = %q, want parent post", posts[0].Record.Text)
	}
}

func TestClient_GetPosts_ChunksRequests(t *testing.T) {
	t.Parallel()

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		uris := r.URL.Query()["uris"]
		if requestCount == 1 && len(uris) != 25 {
			t.Fatalf("first request len(uris) = %d, want 25", len(uris))
		}
		if requestCount == 2 && len(uris) != 1 {
			t.Fatalf("second request len(uris) = %d, want 1", len(uris))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"posts":[]}`))
	}))
	t.Cleanup(server.Close)

	client := bluesky.NewClientWith(server.URL+"/xrpc", server.Client())

	uris := make([]string, 26)
	for i := range uris {
		uris[i] = "at://did:plc:example/app.bsky.feed.post/p" + strconv.Itoa(i)
	}

	if _, err := client.GetPosts(context.Background(), uris); err != nil {
		t.Fatalf("GetPosts() err = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("requestCount = %d, want 2", requestCount)
	}
}

func TestPost_AllLabels_IncludesViewAndSelfLabels(t *testing.T) {
	t.Parallel()

	var post bluesky.Post
	if err := json.Unmarshal([]byte(`{
		"uri": "at://did:plc:author/app.bsky.feed.post/abc",
		"author": {"did": "did:plc:author", "handle": "author.example"},
		"record": {
			"text": "hello",
			"createdAt": "2026-01-15T12:00:00.000Z",
			"labels": {
				"$type": "com.atproto.label.defs#selfLabels",
				"values": [{"val": "nudity"}]
			}
		},
		"labels": [{"val": "sexual", "src": "did:plc:moderation"}]
	}`), &post); err != nil {
		t.Fatalf("Unmarshal() err = %v", err)
	}

	labels := post.AllLabels()
	if len(labels) != 2 {
		t.Fatalf("len(labels) = %d, want 2", len(labels))
	}
	if labels[0].Val != "sexual" || labels[0].Src != "did:plc:moderation" {
		t.Fatalf("labels[0] = %#v, want sexual from moderation labeler", labels[0])
	}
	if labels[1].Val != "nudity" || labels[1].Src != "did:plc:author" {
		t.Fatalf("labels[1] = %#v, want nudity self-label", labels[1])
	}
}
