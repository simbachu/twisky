package home_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query/home"
	"github.com/simbachu/twisky/internal/response"
)

type stubReader struct {
	timelineResp *bluesky.AuthorFeedResponse
	feedResp     *bluesky.AuthorFeedResponse
	savedFeeds   []bluesky.SavedFeed
	generators   []bluesky.FeedGenerator
	err          error
	savedErr     error
	generatorsErr error
	profiles     []bluesky.Profile
	parentPosts  []bluesky.Post

	lastTimelineRequest bluesky.TimelineRequest
	lastFeedRequest     bluesky.FeedRequest
	lastGetPostsURIs    []string
}

func (s *stubReader) GetTimeline(_ context.Context, req bluesky.TimelineRequest) (*bluesky.AuthorFeedResponse, error) {
	s.lastTimelineRequest = req
	if s.err != nil {
		return nil, s.err
	}
	return s.timelineResp, nil
}

func (s *stubReader) GetSavedFeeds(context.Context) ([]bluesky.SavedFeed, error) {
	if s.savedErr != nil {
		return nil, s.savedErr
	}
	if s.err != nil {
		return nil, s.err
	}
	return s.savedFeeds, nil
}

func (s *stubReader) GetFeedGenerators(context.Context, []string) ([]bluesky.FeedGenerator, error) {
	if s.generatorsErr != nil {
		return nil, s.generatorsErr
	}
	if s.err != nil {
		return nil, s.err
	}
	return s.generators, nil
}

func (s *stubReader) GetFeed(_ context.Context, req bluesky.FeedRequest) (*bluesky.AuthorFeedResponse, error) {
	s.lastFeedRequest = req
	if s.err != nil {
		return nil, s.err
	}
	return s.feedResp, nil
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

func TestHandler_Handle_HeadCheckSkipsParentFetch(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		timelineResp: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/abc",
					Author: bluesky.Author{Handle: "dev.example"},
					Record: bluesky.PostRecord{
						Text: "reply",
						Reply: &bluesky.RecordReplyRef{
							Parent: bluesky.StrongRef{URI: "at://did:plc:example/app.bsky.feed.post/parent"},
							Root:   bluesky.StrongRef{URI: "at://did:plc:example/app.bsky.feed.post/parent"},
						},
					},
				},
			}},
		},
		parentPosts: []bluesky.Post{{
			URI:    "at://did:plc:example/app.bsky.feed.post/parent",
			Author: bluesky.Author{Handle: "dev.example"},
			Record: bluesky.PostRecord{Text: "parent"},
		}},
	}
	handler := home.NewHandler(reader, nil)

	resp := handler.Handle(context.Background(), intent.ViewHome{HeadCheck: true})
	view, ok := resp.(home.HomeView)
	if !ok {
		t.Fatalf("Handle() type = %T, want HomeView", resp)
	}
	if len(view.Feed.Posts) != 1 || view.Feed.Posts[0].ID != "abc" {
		t.Fatalf("posts = %#v, want abc", view.Feed.Posts)
	}
	if len(reader.lastGetPostsURIs) != 0 {
		t.Fatalf("GetPosts URIs = %v, want none on HeadCheck", reader.lastGetPostsURIs)
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

func TestHandler_HandleSavedFeedsErrorStillServesFollowing(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		timelineResp: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					URI:    "at://did:plc:example/app.bsky.feed.post/abc",
					Author: bluesky.Author{Handle: "dev.example"},
					Record: bluesky.PostRecord{Text: "timeline post"},
				},
			}},
		},
		savedErr: errors.New("preferences down"),
	}
	resp := home.NewHandler(reader, nil).Handle(context.Background(), intent.ViewHome{})

	view, ok := resp.(home.HomeView)
	if !ok {
		t.Fatalf("Handle() type = %T, want HomeView", resp)
	}
	if len(view.Tabs) != 1 || view.Tabs[0].Label != "Following" {
		t.Fatalf("Tabs = %#v, want Following only", view.Tabs)
	}
	if len(view.Feed.Posts) != 1 || view.Feed.Posts[0].Text != "timeline post" {
		t.Fatalf("Feed.Posts = %#v, want timeline post", view.Feed.Posts)
	}
}

func TestHandler_HandleFeedGeneratorsErrorStillServesFollowing(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		timelineResp: &bluesky.AuthorFeedResponse{},
		savedFeeds: []bluesky.SavedFeed{
			{ID: "one", Pinned: true, Type: "feed", URI: "at://did:plc:one/app.bsky.feed.generator/one"},
		},
		generatorsErr: errors.New("generators down"),
	}
	resp := home.NewHandler(reader, nil).Handle(context.Background(), intent.ViewHome{})

	view, ok := resp.(home.HomeView)
	if !ok {
		t.Fatalf("Handle() type = %T, want HomeView", resp)
	}
	if len(view.Tabs) != 1 || view.Tabs[0].Label != "Following" {
		t.Fatalf("Tabs = %#v, want Following only", view.Tabs)
	}
}

func TestHandler_HandleBuildsPinnedFeedTabsInPreferenceOrder(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		timelineResp: &bluesky.AuthorFeedResponse{},
		savedFeeds: []bluesky.SavedFeed{
			{ID: "one", Pinned: true, Type: "feed", URI: "at://did:plc:one/app.bsky.feed.generator/one"},
			{ID: "list", Pinned: true, Type: "list", URI: "at://did:plc:one/app.bsky.graph.list/list"},
			{ID: "two", Pinned: false, Type: "feed", URI: "at://did:plc:two/app.bsky.feed.generator/two"},
			{ID: "three", Pinned: true, Type: "feed", URI: "at://did:plc:three/app.bsky.feed.generator/three"},
		},
		generators: []bluesky.FeedGenerator{
			{URI: "at://did:plc:three/app.bsky.feed.generator/three", DisplayName: "Three"},
			{URI: "at://did:plc:one/app.bsky.feed.generator/one", DisplayName: "One"},
		},
	}

	resp := home.NewHandler(reader, nil).Handle(context.Background(), intent.ViewHome{})

	view, ok := resp.(home.HomeView)
	if !ok {
		t.Fatalf("Handle() type = %T, want HomeView", resp)
	}
	if len(view.Tabs) != 3 {
		t.Fatalf("len(view.Tabs) = %d, want 3", len(view.Tabs))
	}
	if view.Tabs[0].Label != "Following" || !view.Tabs[0].Current {
		t.Fatalf("Tabs[0] = %#v, want current Following", view.Tabs[0])
	}
	if view.Tabs[1].Label != "One" || view.Tabs[2].Label != "Three" {
		t.Fatalf("Tabs = %#v, want Following, One, Three", view.Tabs)
	}
}

func TestHandler_HandleLoadsSelectedGeneratorAndPassesCursor(t *testing.T) {
	t.Parallel()

	uri := "at://did:plc:example/app.bsky.feed.generator/for-you"
	reader := &stubReader{
		savedFeeds: []bluesky.SavedFeed{{Pinned: true, Type: "feed", URI: uri}},
		generators: []bluesky.FeedGenerator{{URI: uri, DisplayName: "For You"}},
		feedResp: &bluesky.AuthorFeedResponse{Feed: []bluesky.FeedItem{{
			Post: bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/custom",
				Author: bluesky.Author{Handle: "dev.example"},
				Record: bluesky.PostRecord{Text: "custom feed post"},
			},
		}}},
	}

	resp := home.NewHandler(reader, nil).Handle(context.Background(), intent.ViewHome{
		FeedSlug: "for-you",
		Cursor:   "cursor-1",
	})

	view, ok := resp.(home.HomeView)
	if !ok {
		t.Fatalf("Handle() type = %T, want HomeView", resp)
	}
	if reader.lastFeedRequest.URI != uri || reader.lastFeedRequest.Cursor != "cursor-1" {
		t.Fatalf("lastFeedRequest = %#v, want URI %q and cursor-1", reader.lastFeedRequest, uri)
	}
	if reader.lastTimelineRequest != (bluesky.TimelineRequest{}) {
		t.Fatalf("lastTimelineRequest = %#v, want zero request", reader.lastTimelineRequest)
	}
	if view.Title != "For You" || view.Path != "/feed/for-you" {
		t.Fatalf("view title/path = %q/%q, want For You//feed/for-you", view.Title, view.Path)
	}
	if !view.Tabs[1].Current {
		t.Fatalf("Tabs[1] = %#v, want current", view.Tabs[1])
	}
}

func TestHandler_HandleDisambiguatesDuplicateFeedNames(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		timelineResp: &bluesky.AuthorFeedResponse{},
		savedFeeds: []bluesky.SavedFeed{
			{Pinned: true, Type: "feed", URI: "at://did:plc:one/app.bsky.feed.generator/discover"},
			{Pinned: true, Type: "feed", URI: "at://did:plc:two/app.bsky.feed.generator/discover"},
		},
		generators: []bluesky.FeedGenerator{
			{URI: "at://did:plc:one/app.bsky.feed.generator/discover", DisplayName: "Discover"},
			{URI: "at://did:plc:two/app.bsky.feed.generator/discover", DisplayName: "Discover"},
		},
	}

	resp := home.NewHandler(reader, nil).Handle(context.Background(), intent.ViewHome{})

	view := resp.(home.HomeView)
	if view.Tabs[1].Slug == view.Tabs[2].Slug {
		t.Fatalf("duplicate slugs = %q, want distinct", view.Tabs[1].Slug)
	}
	for _, tab := range view.Tabs[1:] {
		if !strings.HasPrefix(tab.Slug, "discover-") {
			t.Fatalf("slug = %q, want discover- prefix", tab.Slug)
		}
	}
}

func TestHandler_HandleUnknownFeedSlug(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		savedFeeds: []bluesky.SavedFeed{},
		generators: []bluesky.FeedGenerator{},
	}

	resp := home.NewHandler(reader, nil).Handle(context.Background(), intent.ViewHome{FeedSlug: "missing"})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusNotFound {
		t.Fatalf("Status = %d, want %d", errResp.Status, http.StatusNotFound)
	}
	if errResp.Message != "Could not find feed missing" {
		t.Fatalf("Message = %q, want Could not find feed missing", errResp.Message)
	}
}

func TestHandler_HandleTimelineUpstreamError(t *testing.T) {
	t.Parallel()

	reader := &timelineFailReader{
		stubReader: stubReader{
			savedFeeds: []bluesky.SavedFeed{},
			generators: []bluesky.FeedGenerator{},
		},
		timelineErr: errors.New("timeline down"),
	}
	resp := home.NewHandler(reader, nil).Handle(context.Background(), intent.ViewHome{})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadGateway {
		t.Fatalf("Status = %d, want %d", errResp.Status, http.StatusBadGateway)
	}
	if errResp.Message != "Failed to load feed Following" {
		t.Fatalf("Message = %q, want Failed to load feed Following", errResp.Message)
	}
}

func TestHandler_HandleTimelineInsufficientAuth(t *testing.T) {
	t.Parallel()

	reader := &timelineFailReader{
		stubReader: stubReader{
			savedFeeds: []bluesky.SavedFeed{},
			generators: []bluesky.FeedGenerator{},
		},
		timelineErr: errors.New("failed to refresh OAuth tokens: token refresh failed: auth server request failed (HTTP 400): invalid_grant"),
	}
	resp := home.NewHandler(reader, nil).Handle(context.Background(), intent.ViewHome{})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusUnauthorized {
		t.Fatalf("Status = %d, want %d", errResp.Status, http.StatusUnauthorized)
	}
	if !strings.Contains(errResp.Message, "Log out and sign in again") {
		t.Fatalf("Message = %q, want re-login guidance", errResp.Message)
	}
}

type timelineFailReader struct {
	stubReader
	timelineErr error
}

func (r *timelineFailReader) GetTimeline(ctx context.Context, req bluesky.TimelineRequest) (*bluesky.AuthorFeedResponse, error) {
	if r.timelineErr != nil {
		return nil, r.timelineErr
	}
	return r.stubReader.GetTimeline(ctx, req)
}
