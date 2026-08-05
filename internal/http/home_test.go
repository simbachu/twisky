package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/auth/session"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/command"
	twiskyhttp "github.com/simbachu/twisky/internal/http"
	"github.com/simbachu/twisky/internal/query"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/suggestions"
	"github.com/simbachu/twisky/internal/query/tag"
)

type stubHomeReader struct {
	timeline *bluesky.AuthorFeedResponse
	err      error
}

func (s stubHomeReader) GetTimeline(context.Context, bluesky.TimelineRequest) (*bluesky.AuthorFeedResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.timeline, nil
}

func (s stubHomeReader) GetPosts(context.Context, []string) ([]bluesky.Post, error) {
	return nil, nil
}

func (s stubHomeReader) GetProfiles(context.Context, []string) ([]bluesky.Profile, error) {
	return nil, nil
}

func TestHome_AuthDisabledReturns404(t *testing.T) {
	t.Parallel()

	handler := newTestServer(stubReader{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestHome_LoggedOutRendersLogin(t *testing.T) {
	t.Parallel()

	handler, _ := newAuthTestServer(t, "https://dev.twisky.app")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Login with Bluesky") {
		t.Fatalf("body missing login CTA: %s", body)
	}
	if strings.Contains(body, `aria-label="Site"`) {
		t.Fatalf("logged-out home should not include site nav")
	}
}

func TestHome_LoggedInRendersFollowFeed(t *testing.T) {
	t.Parallel()

	auth, err := authoauth.NewService(authoauth.Config{
		PublicBaseURL: "https://dev.twisky.app",
		SessionSecret: "test-secret-at-least-32-bytes-long!!",
		StorePath:     filepath.Join(t.TempDir(), "oauth.db"),
		SecureCookies: true,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	t.Cleanup(func() { _ = auth.Close() })

	queries := query.NewDispatcher(
		profile.NewHandler(stubReader{}, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil),
	)
	server := twiskyhttp.NewServer(queries, command.NewDispatcher(), suggestions.NewHandler(stubReader{}, nil), "https://dev.twisky.app", auth).
		WithHomeReader(stubHomeReader{
			timeline: &bluesky.AuthorFeedResponse{
				Feed: []bluesky.FeedItem{{
					Post: bluesky.Post{
						URI:    "at://did:plc:example/app.bsky.feed.post/home1",
						CID:    "bafyhome1",
						Author: bluesky.Author{Handle: "friend.example", DisplayName: "Friend"},
						Record: bluesky.PostRecord{Text: "timeline post"},
					},
				}},
			},
		})

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts:  []session.Account{{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(recCookie.Result().Cookies()[0])
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		"timeline post",
		`id="feed-list"`,
		`<a href="/" aria-current="page">`,
		`aria-label="Site"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q; got:\n%s", want, body)
		}
	}
}

func TestHome_DoesNotCaptureAsProfileSlug(t *testing.T) {
	t.Parallel()

	handler, _ := newAuthTestServer(t, "https://dev.twisky.app")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (home route, not profile)", rec.Code)
	}
	if strings.Contains(rec.Body.String(), `class="profile"`) {
		t.Fatalf("GET / should not render a profile page")
	}
}
