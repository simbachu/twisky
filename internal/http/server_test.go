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

func (s stubReader) GetPosts(context.Context, []string) ([]bluesky.Post, error) {
	return nil, nil
}

func (s stubReader) GetPostThread(context.Context, string) (bluesky.ThreadNode, error) {
	if s.threadErr != nil {
		return nil, s.threadErr
	}
	return s.thread, nil
}

func newTestServer(reader stubReader) http.Handler {
	queries := query.NewDispatcher(
		profile.NewHandler(reader, nil),
		tag.NewHandler(reader, nil),
		post.NewHandler(reader, nil),
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
}

func TestHandlePost_AncestorsFragment(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{
			DID:    "did:plc:example",
			Handle: "dev.example",
		},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/reply",
				Author: bluesky.Author{Handle: "dev.example", DisplayName: "Dev"},
				Record: bluesky.PostRecord{Text: "linked reply"},
			},
			Parent: bluesky.ThreadViewPost{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/grandparent",
					Author: bluesky.Author{Handle: "other.example", DisplayName: "Other"},
					Record: bluesky.PostRecord{Text: "grandparent post"},
				},
				Parent: bluesky.ThreadViewPost{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/parent",
						Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
						Record: bluesky.PostRecord{Text: "parent post"},
					},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/dev.example/post/reply?ancestors=1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, `<html`) {
		t.Fatalf("body = %q, want ancestors fragment without page wrapper", body)
	}
	if !strings.Contains(body, "parent post") || !strings.Contains(body, "grandparent post") {
		t.Fatalf("body = %q, want ancestor post text", body)
	}
}

func TestHandlePost_AncestorsQueryRequiresOne(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{
			DID:    "did:plc:example",
			Handle: "dev.example",
		},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/reply",
				Author: bluesky.Author{Handle: "dev.example", DisplayName: "Dev"},
				Record: bluesky.PostRecord{Text: "linked reply"},
			},
			Parent: bluesky.ThreadViewPost{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/parent",
					Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
					Record: bluesky.PostRecord{Text: "parent post"},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/dev.example/post/reply?ancestors=0", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `<html`) {
		t.Fatalf("body = %q, want full page for ancestors=0", body)
	}
	if strings.Contains(body, "parent post") {
		t.Fatalf("body = %q, want ancestors omitted on full page", body)
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

func TestHandleStaticStyleCSS_OK(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{})

	req := httptest.NewRequest(http.MethodGet, "/static/styles/style.css", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "--content-width") {
		t.Fatalf("body = %q, want to contain --content-width", body)
	}
}

func TestHandleSlug_CursorFragment(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{
			Handle:      "bsky.app",
			DisplayName: "Bluesky",
		},
		feed: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/page-two",
					Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
					Record: bluesky.PostRecord{Text: "page two post"},
				},
			}},
			Cursor: "page-3",
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app?cursor=page-2", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, "<html") {
		t.Fatalf("body = %q, want fragment without full page wrapper", body)
	}
	if !strings.Contains(body, "page two post") {
		t.Fatalf("body = %q, want page two post", body)
	}
}

func TestHandleSlug_SinceFragment(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{
			Handle:      "bsky.app",
			DisplayName: "Bluesky",
		},
		feed: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{
				{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/new-one",
						Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
						Record: bluesky.PostRecord{Text: "new one"},
					},
				},
				{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/new-two",
						Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
						Record: bluesky.PostRecord{Text: "new two"},
					},
				},
				{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/top",
						Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
						Record: bluesky.PostRecord{Text: "top post"},
					},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app?since=top", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, "<html") {
		t.Fatalf("body = %q, want fragment without full page wrapper", body)
	}
	if !strings.Contains(body, "Show 2 new posts") {
		t.Fatalf("body = %q, want new posts banner", body)
	}
}

func TestHandleSlug_RefreshFragment(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{
			Handle:      "bsky.app",
			DisplayName: "Bluesky",
		},
		feed: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{
				{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/new-one",
						Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
						Record: bluesky.PostRecord{Text: "new one"},
					},
				},
				{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/top",
						Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
						Record: bluesky.PostRecord{Text: "top post"},
					},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app?refresh=top", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, "<html") {
		t.Fatalf("body = %q, want fragment without full page wrapper", body)
	}
	if !strings.Contains(body, "new one") {
		t.Fatalf("body = %q, want prepended post", body)
	}
	if strings.Contains(body, "top post") {
		t.Fatalf("body = %q, want top post omitted from prepend fragment", body)
	}
}

func TestHandleTagged_CursorFragment(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		searchResp: &bluesky.SearchPostsResponse{
			Posts: []bluesky.Post{{
				URI:    "at://did:plc:example/app.bsky.feed.post/page-two",
				Author: bluesky.Author{Handle: "dev.example", DisplayName: "Developer"},
				Record: bluesky.PostRecord{Text: "tag page two"},
			}},
			Cursor: "page-3",
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/tagged/golang?cursor=page-2", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, "<html") {
		t.Fatalf("body = %q, want fragment without full page wrapper", body)
	}
	if !strings.Contains(body, "tag page two") {
		t.Fatalf("body = %q, want tag page two", body)
	}
}
