package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
	twiskyhttp "github.com/simbachu/twisky/internal/http"
	"github.com/simbachu/twisky/internal/query"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/suggestions"
	"github.com/simbachu/twisky/internal/query/tag"
	"github.com/simbachu/twisky/internal/version"
)

type stubReader struct {
	profile    *bluesky.Profile
	feed       *bluesky.AuthorFeedResponse
	searchResp *bluesky.SearchPostsResponse
	thread     bluesky.ThreadNode
	profiles   []bluesky.Profile
	posts      []bluesky.Post
	err        error
	feedErr    error
	searchErr  error
	threadErr  error
	postsErr   error
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
	if s.postsErr != nil {
		return nil, s.postsErr
	}
	return s.posts, nil
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
	return twiskyhttp.NewServer(queries, suggestions.NewHandler(reader, nil), "https://twisky.test").Handler()
}

func TestHandleHealthz_OK(t *testing.T) {
	t.Parallel()

	prev := version.BuildID
	t.Cleanup(func() { version.BuildID = prev })
	version.BuildID = "9c8a405abcdef"

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	newTestServer(stubReader{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if body != "ok 9c8a405" {
		t.Fatalf("body = %q, want %q", body, "ok 9c8a405")
	}
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
	if !strings.Contains(body, `property="og:title" content="Bluesky (@bsky.app)"`) {
		t.Fatalf("body = %q, want og:title", body)
	}
	if !strings.Contains(body, `property="og:url" content="https://twisky.test/bsky.app"`) {
		t.Fatalf("body = %q, want og:url", body)
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

func TestHandlePost_CountsFragment_InitialPollIncludesAllSpans(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		posts: []bluesky.Post{{
			URI:         "at://did:plc:example/app.bsky.feed.post/root",
			Author:      bluesky.Author{Handle: "bsky.app"},
			Record:      bluesky.PostRecord{Text: "root post"},
			LikeCount:   42,
			RepostCount: 3,
			ReplyCount:  1,
		}},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/root?counts=1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", got)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`id="like-count-root"`,
		`id="reply-count-root"`,
		`id="repost-count-root"`,
		`hx-swap-oob="true"`,
		`id="counts-announcer-root"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %s", body, want)
		}
	}
}

func TestHandlePost_CountsFragment_OmitsUnchangedSpans(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		posts: []bluesky.Post{{
			URI:         "at://did:plc:example/app.bsky.feed.post/root",
			Author:      bluesky.Author{Handle: "bsky.app"},
			Record:      bluesky.PostRecord{Text: "root post"},
			LikeCount:   42,
			RepostCount: 3,
			ReplyCount:  1,
		}},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/root?counts=1&like=42&reply=1&repost=3", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "" {
		t.Fatalf("body = %q, want empty response when nothing changed", body)
	}
}

func TestHandlePost_CountsFragment_IncludesOnlyChangedSpan(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		posts: []bluesky.Post{{
			URI:         "at://did:plc:example/app.bsky.feed.post/root",
			Author:      bluesky.Author{Handle: "bsky.app"},
			Record:      bluesky.PostRecord{Text: "root post"},
			LikeCount:   42,
			RepostCount: 3,
			ReplyCount:  1,
		}},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/root?counts=1&like=41&reply=1&repost=3", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="like-count-root"`) {
		t.Fatalf("body = %q, want the changed like span", body)
	}
	if strings.Contains(body, `id="reply-count-root"`) || strings.Contains(body, `id="repost-count-root"`) {
		t.Fatalf("body = %q, want unchanged reply/repost spans omitted", body)
	}
}

func TestHandlePost_CountsFragment_LiveToggleOn(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		posts: []bluesky.Post{{
			URI:    "at://did:plc:example/app.bsky.feed.post/root",
			Author: bluesky.Author{Handle: "bsky.app"},
			Record: bluesky.PostRecord{Text: "root post"},
		}},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/root?counts=1&live=1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`aria-pressed="true"`,
		`aria-label="Pause live counts"`,
		`data-live="true"`,
		`data-counts-poll`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %s", body, want)
		}
	}
}

func TestHandlePost_CountsFragment_LiveToggleOff(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		posts: []bluesky.Post{{
			URI:    "at://did:plc:example/app.bsky.feed.post/root",
			Author: bluesky.Author{Handle: "bsky.app"},
			Record: bluesky.PostRecord{Text: "root post"},
		}},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/root?counts=1&live=0", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`aria-pressed="false"`,
		`aria-label="Show live counts"`,
		`data-live="false"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %s", body, want)
		}
	}
	if strings.Contains(body, "data-href") {
		t.Fatalf("body = %q, want no scheduler data-href once paused", body)
	}
}

func TestHandlePost_FullPage_ExplicitLiveQueryParamStartsOldPostLive(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app"},
				Record: bluesky.PostRecord{Text: "root post", CreatedAt: time.Now().Add(-48 * time.Hour)},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/root?live=1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `aria-pressed="true"`) {
		t.Fatalf("body = %q, want the toggle pre-armed live via ?live=1", body)
	}
}

func TestHandlePost_FullPage_OldPostDefaultsToPaused(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app"},
				Record: bluesky.PostRecord{Text: "root post", CreatedAt: time.Now().Add(-48 * time.Hour)},
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
	if !strings.Contains(body, `aria-pressed="false"`) {
		t.Fatalf("body = %q, want the toggle paused by default for an old post", body)
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

func TestHandlePost_RepliesFragment_EmptyWhenAllKnown(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app"},
				Record: bluesky.PostRecord{Text: "root post"},
			},
			Replies: []bluesky.ThreadNode{
				bluesky.ThreadViewPost{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/reply1",
						Author: bluesky.Author{Handle: "dev.example"},
						Record: bluesky.PostRecord{Text: "reply one"},
					},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/root?replies=1&known=reply1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "" {
		t.Fatalf("body = %q, want empty when all replies are known", body)
	}
}

func TestHandlePost_RepliesFragment_SwapsWhenUnknown(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app"},
				Record: bluesky.PostRecord{Text: "root post"},
			},
			Replies: []bluesky.ThreadNode{
				bluesky.ThreadViewPost{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/reply1",
						Author: bluesky.Author{Handle: "dev.example"},
						Record: bluesky.PostRecord{Text: "reply one"},
					},
				},
				bluesky.ThreadViewPost{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/reply2",
						Author: bluesky.Author{Handle: "dev.example"},
						Record: bluesky.PostRecord{Text: "reply two"},
					},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/root?replies=1&known=reply1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`id="post-replies-root"`,
		`hx-swap-oob="true"`,
		`id="post-reply2"`,
		"reply two",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %s", body, want)
		}
	}
}

func TestHandlePost_FullPage_IncludesEmptyRepliesContainer(t *testing.T) {
	t.Parallel()

	server := newTestServer(stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app"},
				Record: bluesky.PostRecord{Text: "root post", CreatedAt: time.Now()},
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
	if !strings.Contains(body, `id="post-replies-root"`) {
		t.Fatalf("body = %q, want empty replies container", body)
	}
	if !strings.Contains(body, `data-replies-href="/bsky.app/post/root?replies=1"`) {
		t.Fatalf("body = %q, want live poller replies href", body)
	}
}
