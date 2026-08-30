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

type stubSettingsClient struct {
	profile  *bluesky.Profile
	prefs    bluesky.Preferences
	feeds    []bluesky.FeedGenerator
	put      bluesky.Preferences
	putCalls int
	putErr   error
}

func (s *stubSettingsClient) GetProfile(context.Context, string) (*bluesky.Profile, error) {
	return s.profile, nil
}

func (s *stubSettingsClient) GetPreferences(context.Context) (bluesky.Preferences, error) {
	return s.prefs, nil
}

func (s *stubSettingsClient) GetFeedGenerators(context.Context, []string) ([]bluesky.FeedGenerator, error) {
	return s.feeds, nil
}

func (s *stubSettingsClient) PutPreferences(_ context.Context, prefs bluesky.Preferences) error {
	s.putCalls++
	s.put = prefs
	return s.putErr
}

func newSettingsTestServer(t *testing.T, client *stubSettingsClient) (http.Handler, *http.Cookie) {
	t.Helper()
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
		profile.NewHandler(stubReader{}, nil, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil, nil),
	)
	if client.prefs.Labels == nil {
		prefs, err := bluesky.ParsePreferences(nil)
		if err != nil {
			t.Fatalf("ParsePreferences() err = %v", err)
		}
		client.prefs = prefs
	}
	server := twiskyhttp.NewServer(
		queries,
		command.NewDispatcher(),
		suggestions.NewHandler(stubReader{}, nil, nil),
		"https://dev.twisky.app",
		auth,
		nil,
		nil,
	).WithSettingsReader(client).WithSettingsWriter(client)

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts:  []session.Account{{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return server.Handler(), recCookie.Result().Cookies()[0]
}

func TestSettings_AuthDisabledReturns404(t *testing.T) {
	t.Parallel()

	handler := newTestServer(stubReader{})
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestSettings_LoggedOutRendersLogin(t *testing.T) {
	t.Parallel()

	handler, _ := newAuthTestServer(t, "https://dev.twisky.app")
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Login with Bluesky") {
		t.Fatalf("body missing login page")
	}
}

func TestSettings_LoggedInRendersBlueskyPrefs(t *testing.T) {
	t.Parallel()

	prefs, err := bluesky.ParsePreferences(nil)
	if err != nil {
		t.Fatalf("ParsePreferences() err = %v", err)
	}
	prefs.AdultContentEnabled = true
	prefs.SavedFeeds = []bluesky.SavedFeed{{
		Pinned: true, Type: "feed", URI: "at://did:plc:feeds/app.bsky.feed.generator/for-you",
	}}
	handler, cookie := newSettingsTestServer(t, &stubSettingsClient{
		profile: &bluesky.Profile{DID: "did:plc:alice", Handle: "alice.test", DisplayName: "Alice"},
		prefs:   prefs,
		feeds:   []bluesky.FeedGenerator{{URI: "at://did:plc:feeds/app.bsky.feed.generator/for-you", DisplayName: "For You"}},
	})

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`<a href="/settings" aria-current="page">`,
		">Settings</h1>",
		"Alice",
		"Enable adult content",
		"For You · Pinned",
		`id="settings-sign-out"`,
		`href="/settings">Settings</a>`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q; got:\n%s", want, body)
		}
	}
}

func TestSettings_ContentFilteringHTMXReturnsSection(t *testing.T) {
	t.Parallel()

	client := &stubSettingsClient{}
	handler, cookie := newSettingsTestServer(t, client)
	req := httptest.NewRequest(http.MethodPost, "/settings/content-filtering", strings.NewReader(
		"adult_content=1&porn=hide&sexual=ignore&nudity=warn&graphic-media=hide",
	))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if client.putCalls != 1 {
		t.Fatalf("putCalls = %d, want 1", client.putCalls)
	}
	if !client.put.AdultContentEnabled {
		t.Fatal("PutPreferences adult content not enabled")
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="settings-content-filtering"`) {
		t.Fatalf("body = %q, want content-filtering section", body)
	}
	if strings.Contains(body, "<html") {
		t.Fatalf("htmx response should be a fragment")
	}
}

func TestSettings_ThreadingRedirectsWithoutHTMX(t *testing.T) {
	t.Parallel()

	client := &stubSettingsClient{}
	handler, cookie := newSettingsTestServer(t, client)
	req := httptest.NewRequest(http.MethodPost, "/settings/threading", strings.NewReader(
		"sort=oldest&prioritize_followed=1",
	))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303; body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Location") != "/settings" {
		t.Fatalf("Location = %q, want /settings", rec.Header().Get("Location"))
	}
	if client.putCalls != 1 {
		t.Fatalf("putCalls = %d, want 1", client.putCalls)
	}
	if client.put.ThreadView.Sort != bluesky.ThreadSortOldest {
		t.Fatalf("sort = %q, want oldest", client.put.ThreadView.Sort)
	}
}
