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
	postquery "github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/suggestions"
	"github.com/simbachu/twisky/internal/query/tag"
)

type trackingPostReader struct {
	stubReader
	getPostThreadCalls int
}

func (t *trackingPostReader) GetPostThread(ctx context.Context, postURI string, opts *bluesky.PostThreadOpts) (bluesky.ThreadNode, error) {
	t.getPostThreadCalls++
	return t.stubReader.GetPostThread(ctx, postURI, opts)
}

func newPostPageAuthServer(t *testing.T, reader postquery.Reader) (http.Handler, *http.Cookie) {
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
		postquery.NewHandler(stubReader{}, nil, nil),
	)
	server := twiskyhttp.NewServer(
		queries,
		command.NewDispatcher(),
		suggestions.NewHandler(stubReader{}, nil, nil),
		"https://dev.twisky.app",
		auth,
		nil,
		nil,
	).WithPostPageReader(reader)

	recCookie := httptest.NewRecorder()
	if err := auth.Jar.Save(recCookie, session.State{
		ActiveDID: "did:plc:alice",
		Accounts:  []session.Account{{DID: "did:plc:alice", SessionID: "s1", Handle: "alice.test"}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return server.Handler(), recCookie.Result().Cookies()[0]
}

func TestHandlePost_UsesPostPageReaderOverride(t *testing.T) {
	t.Parallel()

	reader := &trackingPostReader{stubReader: stubReader{
		profile: &bluesky.Profile{DID: "did:plc:example", Handle: "bsky.app"},
		thread: bluesky.ThreadViewPost{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/root",
				Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
				Record: bluesky.PostRecord{Text: "root post"},
			},
		},
	}}
	handler, cookie := newPostPageAuthServer(t, reader)

	req := httptest.NewRequest(http.MethodGet, "/bsky.app/post/root", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if reader.getPostThreadCalls != 1 {
		t.Fatalf("GetPostThread calls = %d, want 1", reader.getPostThreadCalls)
	}
	if !strings.Contains(rec.Body.String(), "root post") {
		t.Fatalf("body missing post text")
	}
}
