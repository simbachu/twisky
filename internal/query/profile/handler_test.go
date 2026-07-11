package profile_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/response"
)

type stubReader struct {
	profile *bluesky.Profile
	feed    *bluesky.AuthorFeedResponse
	err     error
	feedErr error

	profiles []bluesky.Profile

	lastFeedRequest bluesky.AuthorFeedRequest
}

func (s stubReader) GetProfile(context.Context, string) (*bluesky.Profile, error) {
	return s.profile, s.err
}

func (s *stubReader) GetAuthorFeed(_ context.Context, req bluesky.AuthorFeedRequest) (*bluesky.AuthorFeedResponse, error) {
	s.lastFeedRequest = req
	if s.feedErr != nil {
		return nil, s.feedErr
	}
	return s.feed, nil
}

func (s stubReader) GetProfiles(context.Context, []string) ([]bluesky.Profile, error) {
	return s.profiles, nil
}

func (s stubReader) GetPosts(context.Context, []string) ([]bluesky.Post, error) {
	return nil, nil
}

func TestHandler_Handle(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{
			DID:         "did:plc:example",
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
	}
	handler := profile.NewHandler(reader, nil)

	resp := handler.Handle(context.Background(), intent.ViewProfile{Slug: "bsky.app", Tab: intent.ProfileTabPosts})

	view, ok := resp.(profile.ProfileView)
	if !ok {
		t.Fatalf("Handle() type = %T, want ProfileView", resp)
	}
	if view.Handle != "bsky.app" {
		t.Fatalf("Handle() view.Handle = %q, want bsky.app", view.Handle)
	}
	if view.Tab != profile.TabPosts {
		t.Fatalf("Handle() view.Tab = %q, want %q", view.Tab, profile.TabPosts)
	}
	if reader.lastFeedRequest.Filter != bluesky.FilterPostsNoReplies {
		t.Fatalf("lastFeedRequest.Filter = %q, want %q", reader.lastFeedRequest.Filter, bluesky.FilterPostsNoReplies)
	}
	if reader.lastFeedRequest.Limit != profile.ProfileFeedLimit {
		t.Fatalf("lastFeedRequest.Limit = %d, want %d", reader.lastFeedRequest.Limit, profile.ProfileFeedLimit)
	}
	if len(view.Feed.Posts) != 1 {
		t.Fatalf("len(view.Feed.Posts) = %d, want 1", len(view.Feed.Posts))
	}
	if view.Feed.Posts[0].Text != "hello world" {
		t.Fatalf("view.Feed.Posts[0].Text = %q, want hello world", view.Feed.Posts[0].Text)
	}
}

func TestHandler_HandleMediaTab(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{Handle: "bsky.app"},
		feed: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					Author: bluesky.Author{Handle: "bsky.app"},
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
	}
	handler := profile.NewHandler(reader, nil)

	resp := handler.Handle(context.Background(), intent.ViewProfile{Slug: "bsky.app", Tab: intent.ProfileTabMedia})

	view, ok := resp.(profile.ProfileView)
	if !ok {
		t.Fatalf("Handle() type = %T, want ProfileView", resp)
	}
	if view.Tab != profile.TabMedia {
		t.Fatalf("Handle() view.Tab = %q, want %q", view.Tab, profile.TabMedia)
	}
	if reader.lastFeedRequest.Filter != bluesky.FilterPostsWithMedia {
		t.Fatalf("lastFeedRequest.Filter = %q, want %q", reader.lastFeedRequest.Filter, bluesky.FilterPostsWithMedia)
	}
	if len(view.Feed.Posts[0].Images) != 1 {
		t.Fatalf("len(view.Feed.Posts[0].Images) = %d, want 1", len(view.Feed.Posts[0].Images))
	}
	image := view.Feed.Posts[0].Images[0]
	if image.Thumb != "https://example.com/thumb.jpg" {
		t.Fatalf("image.Thumb = %q, want https://example.com/thumb.jpg", image.Thumb)
	}
	if image.Fullsize != "https://example.com/full.jpg" {
		t.Fatalf("image.Fullsize = %q, want https://example.com/full.jpg", image.Fullsize)
	}
	if image.Width != 4000 || image.Height != 3000 {
		t.Fatalf("image dimensions = %dx%d, want 4000x3000", image.Width, image.Height)
	}
}

func TestHandler_HandlePassesNextCursor(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{Handle: "bsky.app"},
		feed: &bluesky.AuthorFeedResponse{
			Feed:   []bluesky.FeedItem{},
			Cursor: "next-page",
		},
	}
	handler := profile.NewHandler(reader, nil)

	resp := handler.Handle(context.Background(), intent.ViewProfile{Slug: "bsky.app", Tab: intent.ProfileTabPosts})

	view, ok := resp.(profile.ProfileView)
	if !ok {
		t.Fatalf("Handle() type = %T, want ProfileView", resp)
	}
	if view.Feed.NextCursor != "next-page" {
		t.Fatalf("view.Feed.NextCursor = %q, want next-page", view.Feed.NextCursor)
	}
}

func TestHandler_HandlePassesCursor(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{Handle: "bsky.app"},
		feed:    &bluesky.AuthorFeedResponse{},
	}
	handler := profile.NewHandler(reader, nil)

	resp := handler.Handle(context.Background(), intent.ViewProfile{
		Slug:   "bsky.app",
		Tab:    intent.ProfileTabPosts,
		Cursor: "page-2",
	})

	if _, ok := resp.(profile.ProfileView); !ok {
		t.Fatalf("Handle() type = %T, want ProfileView", resp)
	}
	if reader.lastFeedRequest.Cursor != "page-2" {
		t.Fatalf("lastFeedRequest.Cursor = %q, want page-2", reader.lastFeedRequest.Cursor)
	}
}

func TestHandler_HandleInvalidSlug(t *testing.T) {
	t.Parallel()

	handler := profile.NewHandler(&stubReader{}, nil)

	resp := handler.Handle(context.Background(), intent.ViewProfile{Slug: "hello", Tab: intent.ProfileTabPosts})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadRequest {
		t.Fatalf("Handle() status = %d, want %d", errResp.Status, http.StatusBadRequest)
	}
}

func TestHandler_HandleNotFound(t *testing.T) {
	t.Parallel()

	handler := profile.NewHandler(&stubReader{err: bluesky.ErrNotFound}, nil)

	resp := handler.Handle(context.Background(), intent.ViewProfile{Slug: "missing.example", Tab: intent.ProfileTabPosts})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusNotFound {
		t.Fatalf("Handle() status = %d, want %d", errResp.Status, http.StatusNotFound)
	}
}

func TestHandler_HandleUpstreamError(t *testing.T) {
	t.Parallel()

	handler := profile.NewHandler(&stubReader{err: errors.New("network failure")}, nil)

	resp := handler.Handle(context.Background(), intent.ViewProfile{Slug: "bsky.app", Tab: intent.ProfileTabPosts})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadGateway {
		t.Fatalf("Handle() status = %d, want %d", errResp.Status, http.StatusBadGateway)
	}
}

func TestHandler_HandleFeedUpstreamError(t *testing.T) {
	t.Parallel()

	handler := profile.NewHandler(&stubReader{
		profile: &bluesky.Profile{Handle: "bsky.app"},
		feedErr: errors.New("network failure"),
	}, nil)

	resp := handler.Handle(context.Background(), intent.ViewProfile{Slug: "bsky.app", Tab: intent.ProfileTabPosts})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadGateway {
		t.Fatalf("Handle() status = %d, want %d", errResp.Status, http.StatusBadGateway)
	}
}

func TestHandler_HandleResolvesMentionHandles(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{Handle: "bsky.app"},
		feed: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					Author: bluesky.Author{Handle: "bsky.app"},
					Record: bluesky.PostRecord{
						Text: "@dev.example hello",
						Facets: []bluesky.Facet{{
							Index: bluesky.FacetIndex{ByteStart: 0, ByteEnd: 12},
							Features: []bluesky.FacetFeature{{
								Type: "app.bsky.richtext.facet#mention",
								DID:  "did:plc:example",
							}},
						}},
					},
				},
			}},
		},
		profiles: []bluesky.Profile{{
			DID:    "did:plc:example",
			Handle: "dev.example",
		}},
	}
	handler := profile.NewHandler(reader, nil)

	resp := handler.Handle(context.Background(), intent.ViewProfile{Slug: "bsky.app", Tab: intent.ProfileTabPosts})

	view, ok := resp.(profile.ProfileView)
	if !ok {
		t.Fatalf("Handle() type = %T, want ProfileView", resp)
	}
	if len(view.Feed.Posts) != 1 {
		t.Fatalf("len(view.Feed.Posts) = %d, want 1", len(view.Feed.Posts))
	}
	segment := view.Feed.Posts[0].TextSegments[0]
	if segment.Mention != "dev.example" {
		t.Fatalf("mention = %q, want dev.example", segment.Mention)
	}
}

func TestHandler_HandlePreservesRepostMetadata(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		profile: &bluesky.Profile{Handle: "reposter.example"},
		feed: &bluesky.AuthorFeedResponse{
			Feed: []bluesky.FeedItem{{
				Post: bluesky.Post{
					Author: bluesky.Author{Handle: "original.example", DisplayName: "Original"},
					Record: bluesky.PostRecord{Text: "original post"},
				},
				Reason: &bluesky.FeedReason{
					RepostedBy: bluesky.Author{Handle: "reposter.example", DisplayName: "Reposter"},
				},
			}},
		},
	}
	handler := profile.NewHandler(reader, nil)

	resp := handler.Handle(context.Background(), intent.ViewProfile{Slug: "reposter.example", Tab: intent.ProfileTabPosts})

	view, ok := resp.(profile.ProfileView)
	if !ok {
		t.Fatalf("Handle() type = %T, want ProfileView", resp)
	}
	if view.Feed.Posts[0].RepostedByMaybe == nil {
		t.Fatal("RepostedByMaybe = nil, want reposter metadata")
	}
	if view.Feed.Posts[0].RepostedByMaybe.Handle != "reposter.example" {
		t.Fatalf("RepostedByMaybe.Handle = %q, want reposter.example", view.Feed.Posts[0].RepostedByMaybe.Handle)
	}
}
