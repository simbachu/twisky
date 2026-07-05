package tag_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query/tag"
	"github.com/simbachu/twisky/internal/response"
)

type stubReader struct {
	searchResp *bluesky.SearchPostsResponse
	err        error
	profiles   []bluesky.Profile

	lastSearchRequest bluesky.SearchPostsRequest
}

func (s *stubReader) SearchPosts(_ context.Context, req bluesky.SearchPostsRequest) (*bluesky.SearchPostsResponse, error) {
	s.lastSearchRequest = req
	if s.err != nil {
		return nil, s.err
	}
	return s.searchResp, nil
}

func (s *stubReader) GetProfiles(context.Context, []string) ([]bluesky.Profile, error) {
	return s.profiles, nil
}

func TestHandler_Handle(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		searchResp: &bluesky.SearchPostsResponse{
			Posts: []bluesky.Post{{
				Author: bluesky.Author{Handle: "dev.example", DisplayName: "Developer"},
				Record: bluesky.PostRecord{Text: "hello #golang"},
			}},
		},
	}
	handler := tag.NewHandler(reader)

	resp := handler.Handle(context.Background(), intent.ViewTag{Tag: "golang"})

	view, ok := resp.(tag.TagView)
	if !ok {
		t.Fatalf("Handle() type = %T, want TagView", resp)
	}
	if view.Tag != "golang" {
		t.Fatalf("Handle() view.Tag = %q, want golang", view.Tag)
	}
	if reader.lastSearchRequest.Tag != "golang" {
		t.Fatalf("lastSearchRequest.Tag = %q, want golang", reader.lastSearchRequest.Tag)
	}
	if reader.lastSearchRequest.Limit != tag.TagFeedLimit {
		t.Fatalf("lastSearchRequest.Limit = %d, want %d", reader.lastSearchRequest.Limit, tag.TagFeedLimit)
	}
	if len(view.Feed.Posts) != 1 {
		t.Fatalf("len(view.Feed.Posts) = %d, want 1", len(view.Feed.Posts))
	}
	if view.Feed.Posts[0].Text != "hello #golang" {
		t.Fatalf("view.Feed.Posts[0].Text = %q, want hello #golang", view.Feed.Posts[0].Text)
	}
}

func TestHandler_HandlePassesNextCursor(t *testing.T) {
	t.Parallel()

	reader := &stubReader{
		searchResp: &bluesky.SearchPostsResponse{
			Posts:  []bluesky.Post{},
			Cursor: "next-page",
		},
	}
	handler := tag.NewHandler(reader)

	resp := handler.Handle(context.Background(), intent.ViewTag{Tag: "golang"})

	view, ok := resp.(tag.TagView)
	if !ok {
		t.Fatalf("Handle() type = %T, want TagView", resp)
	}
	if view.Feed.NextCursor != "next-page" {
		t.Fatalf("view.Feed.NextCursor = %q, want next-page", view.Feed.NextCursor)
	}
}

func TestHandler_HandleInvalidTag(t *testing.T) {
	t.Parallel()

	handler := tag.NewHandler(&stubReader{})

	resp := handler.Handle(context.Background(), intent.ViewTag{Tag: ""})

	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusBadRequest {
		t.Fatalf("Handle() status = %d, want %d", errResp.Status, http.StatusBadRequest)
	}
}

func TestHandler_HandleUpstreamError(t *testing.T) {
	t.Parallel()

	handler := tag.NewHandler(&stubReader{err: errors.New("network failure")})

	resp := handler.Handle(context.Background(), intent.ViewTag{Tag: "golang"})

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
		searchResp: &bluesky.SearchPostsResponse{
			Posts: []bluesky.Post{{
				Author: bluesky.Author{Handle: "dev.example"},
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
			}},
		},
		profiles: []bluesky.Profile{{
			DID:    "did:plc:example",
			Handle: "dev.example",
		}},
	}
	handler := tag.NewHandler(reader)

	resp := handler.Handle(context.Background(), intent.ViewTag{Tag: "golang"})

	view, ok := resp.(tag.TagView)
	if !ok {
		t.Fatalf("Handle() type = %T, want TagView", resp)
	}
	segment := view.Feed.Posts[0].TextSegments[0]
	if segment.Mention != "dev.example" {
		t.Fatalf("mention = %q, want dev.example", segment.Mention)
	}
}
