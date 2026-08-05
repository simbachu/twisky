package home_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query/home"
	"github.com/simbachu/twisky/internal/response"
)

type stubReader struct {
	timelineResp *bluesky.AuthorFeedResponse
	err          error
	profiles     []bluesky.Profile
	parentPosts  []bluesky.Post

	lastTimelineRequest bluesky.TimelineRequest
	lastGetPostsURIs    []string
}

func (s *stubReader) GetTimeline(_ context.Context, req bluesky.TimelineRequest) (*bluesky.AuthorFeedResponse, error) {
	s.lastTimelineRequest = req
	if s.err != nil {
		return nil, s.err
	}
	return s.timelineResp, nil
}

func (s *stubReader) GetProfiles(context.Context, []string) ([]bluesky.Profile, error) {
	return s.profiles, nil
}

func (s *stubReader) GetPosts(_ context.Context, uris []string) ([]bluesky.Post, error) {
	s.lastGetPostsURIs = append([]string(nil), uris...)
	return s.parentPosts, nil
}

func TestHandler_Handle(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		timelineResp: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/abc",
					Author: bluesky.Author{Handle: "dev.example", DisplayName: "Developer"},
					Record: bluesky.PostRecord{Text: "hello from timeline"},
				},
			}},
			Cursor: "next-cursor",
		},
	}
	handler := home.NewHandler(reader, nil)

	resp := handler.Handle(context.Background(), intent.ViewHome{})

	view, ok := resp.(home.HomeView)
	if !ok {
		t.Fatalf("Handle() type = %T, want HomeView", resp)
	}
	if reader.lastTimelineRequest.Limit != home.HomeFeedLimit {
		t.Fatalf("lastTimelineRequest.Limit = %d, want %d", reader.lastTimelineRequest.Limit, home.HomeFeedLimit)
	}
	if len(view.Feed.Posts) != 1 {
		t.Fatalf("len(view.Feed.Posts) = %d, want 1", len(view.Feed.Posts))
	}
	if view.Feed.Posts[0].Text != "hello from timeline" {
		t.Fatalf("view.Feed.Posts[0].Text = %q, want hello from timeline", view.Feed.Posts[0].Text)
	}
	if view.Feed.NextCursor != "next-cursor" {
		t.Fatalf("view.Feed.NextCursor = %q, want next-cursor", view.Feed.NextCursor)
	}
}

func TestHandler_HandlePassesCursor(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		timelineResp: &bluesky.AuthorFeedResponse{},
	}
	handler := home.NewHandler(reader, nil)

	_ = handler.Handle(context.Background(), intent.ViewHome{Cursor: "cursor-1"})

	if reader.lastTimelineRequest.Cursor != "cursor-1" {
		t.Fatalf("lastTimelineRequest.Cursor = %q, want cursor-1", reader.lastTimelineRequest.Cursor)
	}
}

func TestHandler_HandleUpstreamError(t *testing.T) {
	t.Parallel()

	handler := home.NewHandler(&stubReader{err: errors.New("boom")}, nil)
	resp := handler.Handle(context.Background(), intent.ViewHome{})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadGateway {
		t.Fatalf("Status = %d, want %d", errResp.Status, http.StatusBadGateway)
	}
}
