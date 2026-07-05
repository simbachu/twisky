package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	twiskyhttp "github.com/simbachu/twisky/internal/http"
	"github.com/simbachu/twisky/internal/query"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/tag"
)

type stubReader struct {
	profile     *bluesky.Profile
	feed        *bluesky.AuthorFeedResponse
	searchResp  *bluesky.SearchPostsResponse
	thread      bluesky.ThreadNode
	profiles    []bluesky.Profile
	err         error
	feedErr     error
	searchErr   error
	threadErr   error
}

func (s stubReader) GetProfile(context.Context, string) (*bluesky.Profile, error) {
	return s.profile, s.err
}

func (s stubReader) GetAuthorFeed(context.Context, bluesky.AuthorFeedRequest) (*bluesky.AuthorFeedResponse, error) {
	if s.feedErr != nil {
		return nil, s.feedErr
	}
	return s.feed, nil
}

func (s stubReader) SearchPosts(context.Context, bluesky.SearchPostsRequest) (*bluesky.SearchPostsResponse, error) {
	if s.searchErr != nil {
		return nil, s.searchErr
	}
	return s.searchResp, nil
}

func (s stubReader) GetProfiles(context.Context, []string) ([]bluesky.Profile, error) {
	return s.profiles, nil
}

func (s stubReader) GetPostThread(context.Context, string) (bluesky.ThreadNode, error) {
	if s.threadErr != nil {
		return nil, s.threadErr
	}
	return s.thread, nil
}

func newTestServer(reader stubReader) http.Handler {
	queries := query.NewDispatcher(
		profile.NewHandler(reader),
		tag.NewHandler(reader),
		post.NewHandler(reader),
	)
	return twiskyhttp.NewServer(queries).Handler()
}

func TestHandleSlug_OK(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{
			Handle:      "bsky.app",
			DisplayName: "Bluesky",
			Followers:   100,
		},
		feed: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
					Record: bluesky.PostRecord{Text: "hello world"},
				},
			}},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Bluesky") {
		t.Fatalf("body = %q, want to contain Bluesky", body)
	}
	if !strings.Contains(body, "hello world") {
		t.Fatalf("body = %q, want to contain hello world", body)
	}
	if !strings.Contains(body, `aria-current="page"`) || !strings.Contains(body, ">Posts<") {
		t.Fatalf("body = %q, want Posts tab marked current", body)
	}
}

func TestHandleSlug_MediaTab(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{
			Handle:      "bsky.app",
			DisplayName: "Bluesky",
		},
		feed: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
					Record: bluesky.PostRecord{Text: "photo post"},
					Embed: &bluesky.Embed{
						Images: []bluesky.EmbedImage{{
							Thumb:    "https://example.com/thumb.jpg",
							Fullsize: "https://example.com/full.jpg",
							Alt:      "a photo",
							AspectRatio: &bluesky.AspectRatio{
								Width:  4000,
								Height: 3000,
							},
						}},
					},
				},
			}},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/media", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "photo post") {
		t.Fatalf("body = %q, want to contain photo post", body)
	}
	if !strings.Contains(body, `width="4000"`) || !strings.Contains(body, `height="3000"`) {
		t.Fatalf("body = %q, want image aspect ratio attributes", body)
	}
	if !strings.Contains(body, `srcset="https://example.com/thumb.jpg 1000w, https://example.com/full.jpg 2000w"`) {
		t.Fatalf("body = %q, want srcset with thumb and fullsize URLs", body)
	}
	if !strings.Contains(body, "<figure>") {
		t.Fatalf("body = %q, want figure wrapper for post images", body)
	}
	if !strings.Contains(body, ">Media<") {
		t.Fatalf("body = %q, want Media tab link", body)
	}
}

func TestHandleSlug_MediaTab_GalleryEmbed(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{
			Handle:      "bsky.app",
			DisplayName: "Bluesky",
		},
		feed: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
					Record: bluesky.PostRecord{Text: "gallery post"},
					Embed: &bluesky.Embed{
						Type: "app.bsky.embed.gallery#view",
						Items: []bluesky.EmbedImage{
							{
								Thumbnail: "https://example.com/thumb1.jpg",
								Fullsize:  "https://example.com/full1.jpg",
								AspectRatio: &bluesky.AspectRatio{
									Width:  1000,
									Height: 800,
								},
							},
							{
								Thumbnail: "https://example.com/thumb2.jpg",
								Fullsize:  "https://example.com/full2.jpg",
								Alt:       "second",
								AspectRatio: &bluesky.AspectRatio{
									Width:  1200,
									Height: 900,
								},
							},
						},
					},
				},
			}},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/media", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "gallery post") {
		t.Fatalf("body = %q, want to contain gallery post", body)
	}
	if !strings.Contains(body, "<figure>") {
		t.Fatalf("body = %q, want figure wrapper for gallery images", body)
	}
	if !strings.Contains(body, `src="https://example.com/thumb1.jpg"`) {
		t.Fatalf("body = %q, want first gallery thumbnail src", body)
	}
	if !strings.Contains(body, `src="https://example.com/thumb2.jpg"`) {
		t.Fatalf("body = %q, want second gallery thumbnail src", body)
	}
	if !strings.Contains(body, `alt="second"`) {
		t.Fatalf("body = %q, want second gallery image alt text", body)
	}
}

func TestHandleSlug_InvalidSlug(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleSlug_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{err: bluesky.ErrNotFound})

	req := httptest.NewRequest(http.MethodGet, "/missing.example", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandleSlug_UpstreamError(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{err: errors.New("network failure")})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}

func TestHandleTagged_OK(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		searchResp: &bluesky.SearchPostsResponse{
			Posts: []bluesky.Post{{
				Author: bluesky.Author{Handle: "dev.example", DisplayName: "Developer"},
				Record: bluesky.PostRecord{
					Text: "hello #golang",
					Facets: []bluesky.Facet{{
						Index: bluesky.FacetIndex{ByteStart: 6, ByteEnd: 13},
						Features: []bluesky.FacetFeature{{
							Type: "app.bsky.richtext.facet#tag",
							Tag:  "golang",
						}},
					}},
				},
			}},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/tagged/golang", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "#golang") {
		t.Fatalf("body = %q, want to contain #golang", body)
	}
	if !strings.Contains(body, `href="/tagged/golang"`) {
		t.Fatalf("body = %q, want hashtag link to /tagged/golang", body)
	}
	if !strings.Contains(body, "hello ") {
		t.Fatalf("body = %q, want to contain hello ", body)
	}
}

func TestHandleSlug_MentionAndLinkLinks(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{
			Handle:      "dev.example",
			DisplayName: "Developer",
		},
		profiles: []bluesky.Profile{{
			DID:    "did:plc:example",
			Handle: "dev.example",
		}},
		feed: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					Author: bluesky.Author{Handle: "dev.example", DisplayName: "Developer"},
					Record: bluesky.PostRecord{
						Text: "@dev.example see https://example.com",
						Facets: []bluesky.Facet{
							{
								Index: bluesky.FacetIndex{ByteStart: 0, ByteEnd: 12},
								Features: []bluesky.FacetFeature{{
									Type: "app.bsky.richtext.facet#mention",
									DID:  "did:plc:example",
								}},
							},
							{
								Index: bluesky.FacetIndex{ByteStart: 17, ByteEnd: 36},
								Features: []bluesky.FacetFeature{{
									Type: "app.bsky.richtext.facet#link",
									URI:  "https://example.com",
								}},
							},
						},
					},
				},
			}},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/dev.example", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `href="/dev.example"`) {
		t.Fatalf("body = %q, want mention link to /dev.example", body)
	}
	if !strings.Contains(body, `href="https://example.com"`) {
		t.Fatalf("body = %q, want external link to https://example.com", body)
	}
	if !strings.Contains(body, `target="_blank"`) {
		t.Fatalf("body = %q, want external link target _blank", body)
	}
}

func TestHandlePost_OK(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{
			DID:    "did:plc:example",
			Handle: "bsky.app",
		},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
				Record: bluesky.PostRecord{Text: "root post"},
			},
			Replies: []bluesky.ThreadNode{
				bluesky.ThreadViewPost{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/reply1",
						Author: bluesky.Author{Handle: "dev.example", DisplayName: "Dev"},
						Record: bluesky.PostRecord{Text: "reply one"},
					},
					Replies: []bluesky.ThreadNode{
						bluesky.ThreadViewPost{
							Post: bluesky.Post{
								URI:    "at://did:plc:example/app.bsky.feed.post/reply2",
								Author: bluesky.Author{Handle: "dev.example", DisplayName: "Dev"},
								Record: bluesky.PostRecord{Text: "nested reply"},
							},
						},
					},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/root", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "root post") {
		t.Fatalf("body = %q, want to contain root post", body)
	}
	if !strings.Contains(body, "reply one") {
		t.Fatalf("body = %q, want to contain reply one", body)
	}
	if !strings.Contains(body, "nested reply") {
		t.Fatalf("body = %q, want to contain nested reply", body)
	}
	if !strings.Contains(body, `href="/dev.example/post/reply1"`) {
		t.Fatalf("body = %q, want reply link to /dev.example/post/reply1", body)
	}
}

func TestHandlePost_InvalidSlug(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{})

	req := httptest.NewRequest(http.MethodGet, "/hello/post/abc", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlePost_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile:   &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		threadErr: bluesky.ErrNotFound,
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/missing", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
