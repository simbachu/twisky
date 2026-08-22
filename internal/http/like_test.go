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

type stubLikeWriter struct {
	calls     int
	err       error
	uri       string
	cid       string
	recordURI string
}

func (s *stubLikeWriter) CreateLike(_ context.Context, uri, cid string) (string, error) {
	s.calls++
	s.uri = uri
	s.cid = cid
	if s.recordURI == "" {
		s.recordURI = "at://did:plc:me/app.bsky.feed.like/like1"
	}
	return s.recordURI, s.err
}

func (s *stubLikeWriter) DeleteLike(context.Context, string) error {
	s.calls++
	return s.err
}

func newLikeTestServer(t *testing.T, writer *stubLikeWriter) http.Handler {
	t.Helper()
	queries := query.NewDispatcher(
		profile.NewHandler(stubReader{}, nil),
		tag.NewHandler(stubReader{}, nil),
		post.NewHandler(stubReader{}, nil),
	)
	server := twiskyhttp.NewServer(queries, command.NewDispatcher(), suggestions.NewHandler(stubReader{}, nil), "https://twisky.test", nil)
	if writer != nil {
		server = server.WithLikeWriter(writer)
	}
	return server.Handler()
}

func TestLike_NoSessionReturns204(t *testing.T) {
	t.Parallel()

	handler := newLikeTestServer(t, nil)
	req := httptest.NewRequest(http.MethodPost, "/action/like", strings.NewReader("uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=3"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestLike_MissingFieldsReturns400(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{}
	handler := newLikeTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/like", strings.NewReader("uri=&cid=bafy"))
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

func TestLike_AuthedReturnsEngagedFragment(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{}
	handler := newLikeTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/like", strings.NewReader("uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=3"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if writer.calls != 1 || writer.uri == "" || writer.cid != "bafy" {
		t.Fatalf("CreateLike calls=%d uri=%q cid=%q", writer.calls, writer.uri, writer.cid)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`ui-icon-engaged`,
		`ui-action--like`,
		`aria-label="Like, 4"`,
		`hx-post="/action/unlike"`,
		`hx-disabled-elt="this"`,
		`hx-sync="this:drop"`,
		`at://did:plc:me/app.bsky.feed.like/like1`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %s", body, want)
		}
	}
}

func TestLike_PostPageReturnsCountIDAndOOBStats(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{}
	handler := newLikeTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/like", strings.NewReader("uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=3&id=root"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`ui-icon-engaged`,
		`id="like-count-root"`,
		`id="like-stats-root"`,
		`hx-swap-oob="true"`,
		`id="counts-announcer-root"`,
		`4 likes`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %s", body, want)
		}
	}
}

func TestLike_DuplicateTreatedAsSuccess(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{err: errors.New("InvalidRequest: Record already exists")}
	handler := newLikeTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/like", strings.NewReader("uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=1"))
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

func TestLike_WriterErrorReturns502(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{err: errors.New("pds down")}
	handler := newLikeTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/like", strings.NewReader("uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}

func TestUnlike_MissingRecordReturns400(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{}
	handler := newLikeTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/unlike", strings.NewReader("record="))
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

func TestUnlike_AuthedReturnsIdleFragment(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{}
	handler := newLikeTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/unlike", strings.NewReader("record=at://did:plc:me/app.bsky.feed.like/like1&uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=4"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if writer.calls != 1 {
		t.Fatalf("calls = %d, want 1", writer.calls)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`hx-post="/action/like"`,
		`aria-label="Like, 3"`,
		`hx-disabled-elt="this"`,
		`hx-sync="this:drop"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, want %s", body, want)
		}
	}
	if strings.Contains(body, `ui-icon-engaged`) {
		t.Fatalf("body = %q, want unengaged", body)
	}
}

func TestUnlike_NotFoundTreatedAsSuccess(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{err: errors.New("RecordNotFound: could not locate record")}
	handler := newLikeTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/unlike", strings.NewReader("record=at://did:plc:me/app.bsky.feed.like/like1&uri=at://did:plc:x/app.bsky.feed.post/a&cid=bafy&count=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if strings.Contains(rec.Body.String(), `ui-icon-engaged`) {
		t.Fatalf("body = %q, want unengaged", rec.Body.String())
	}
}

func TestUnlike_WriterErrorReturns502(t *testing.T) {
	t.Parallel()

	writer := &stubLikeWriter{err: errors.New("pds down")}
	handler := newLikeTestServer(t, writer)
	req := httptest.NewRequest(http.MethodPost, "/action/unlike", strings.NewReader("record=at://did:plc:me/app.bsky.feed.like/like1&count=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}
