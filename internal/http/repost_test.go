package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/command"
	twiskyhttp "github.com/simbachu/twisky/internal/http"
	"github.com/simbachu/twisky/internal/query"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/suggestions"
	"github.com/simbachu/twisky/internal/query/tag"
)

type stubRepostWriter struct {
	calls int
	err   error
	uri   string
	cid   string
}

func (s *stubRepostWriter) CreateRepost(_ context.Context, uri, cid string) error {
	s.calls++
	s.uri = uri
	s.cid = cid
	return s.err
}

func newRepostTestServer(t *testing.T, writer *stubRepostWriter) http.Handler {
	t.Helper()
	queries := query.NewDispatcher(
		profile.NewHandler(stubReader{}, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil),
	)
	server := twiskyhttp.NewServer(queries, command.NewDispatcher(), suggestions.NewHandler(stubReader{}, nil), "https://twisky.test", nil)
	if writer != nil {
		server = server.WithRepostWriter(writer)
	}
	return server.Handler()
}

func TestRepost_NoSessionReturns204(t *testing.T) {
	t.Parallel()

	handler := newRepostTestServer(t, nil)
	req := httptest.NewRequest(http.MethodPost, "/action/repost", strings.NewReader("uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=3"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestRepost_MissingFieldsReturns400(t *testing.T) {
	t.Parallel()

	writer := &stubRepostWriter{}
	handler := newRepostTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/repost", strings.NewReader("uri=&cid=bafy"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if writer.calls != 0 {
		t.Fatalf("calls = %d, want 0", writer.calls)
	}
}

func TestRepost_AuthedReturnsEngagedFragment(t *testing.T) {
	t.Parallel()

	writer := &stubRepostWriter{}
	handler := newRepostTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/repost", strings.NewReader("uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=3"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if writer.calls != 1 || writer.uri == "" || writer.cid != "bafy" {
		t.Fatalf("CreateRepost calls=%d uri=%q cid=%q", writer.calls, writer.uri, writer.cid)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`ui-icon-engaged`,
		`ui-action--repost`,
		`aria-label="Repost, 4"`,
		`hx-post="/action/repost"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %s", body, want)
		}
	}
	if strings.Contains(body, `id="repost-count-`) || strings.Contains(body, `hx-swap-oob`) {
		t.Fatalf("body = %q, want feed fragment without live count ids or OOB stats", body)
	}
}

func TestRepost_PostPageReturnsCountIDAndOOBStats(t *testing.T) {
	t.Parallel()

	writer := &stubRepostWriter{}
	handler := newRepostTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/repost", strings.NewReader("uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=3&id=root"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`ui-icon-engaged`,
		`id="repost-count-root"`,
		`id="repost-stats-root"`,
		`hx-swap-oob="true"`,
		`id="counts-announcer-root"`,
		`4 reposts`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %s", body, want)
		}
	}
}

func TestRepost_DuplicateTreatedAsSuccess(t *testing.T) {
	t.Parallel()

	writer := &stubRepostWriter{err: errors.New("InvalidRequest: Record already exists")}
	handler := newRepostTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/repost", strings.NewReader("uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `ui-icon-engaged`) {
		t.Fatalf("body = %q, want engaged", rec.Body.String())
	}
}

func TestRepost_WriterErrorReturns502(t *testing.T) {
	t.Parallel()

	writer := &stubRepostWriter{err: errors.New("pds down")}
	handler := newRepostTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/repost", strings.NewReader("uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}
