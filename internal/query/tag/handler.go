package tag

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/response"
)

type Reader interface {
	SearchPosts(ctx context.Context, req bluesky.SearchPostsRequest) (*bluesky.SearchPostsResponse, error)
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
}

type Handler struct {
	reader Reader
}

const TagFeedLimit = 20

func NewHandler(reader Reader) *Handler {
	return &Handler{reader: reader}
}

// TagView is the read model returned for a hashtag page.
type TagView struct {
	Tag  string
	Feed feedquery.FeedView
}

func (TagView) IsResponse() {}

func (h *Handler) Handle(ctx context.Context, i intent.ViewTag) response.Response {
	tag := strings.TrimSpace(i.Tag)
	if tag == "" {
		return response.ErrorResponse{Status: http.StatusBadRequest, Message: "invalid tag"}
	}

	items, err := h.reader.SearchPosts(ctx, bluesky.SearchPostsRequest{
		Tag:   tag,
		Limit: TagFeedLimit,
	})
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "tag not found"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	return TagView{
		Tag:  tag,
		Feed: feedquery.ResolveMentionHandles(ctx, h.reader, feedquery.NewFeedView(items.Posts, items.Cursor)),
	}
}
