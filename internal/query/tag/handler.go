package tag

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/moderation"
	"github.com/simbachu/twisky/internal/response"
)

type Reader interface {
	SearchPosts(ctx context.Context, req bluesky.SearchPostsRequest) (*bluesky.SearchPostsResponse, error)
	GetPosts(ctx context.Context, uris []string) ([]bluesky.Post, error)
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
}

type Handler struct {
	reader Reader
	prefs  moderation.PrefsProvider
}

const TagFeedLimit = 20

func NewHandler(reader Reader, prefs moderation.PrefsProvider) *Handler {
	if prefs == nil {
		prefs = moderation.DefaultPrefsProvider{}
	}
	return &Handler{reader: reader, prefs: prefs}
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
		Tag:    tag,
		Limit:  TagFeedLimit,
		Cursor: i.Cursor,
	})
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "tag not found"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	feed := feedquery.NewFeedView(items.Posts, items.Cursor)
	feed, err = feedquery.EnrichReplyParents(ctx, h.reader, feed)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	return TagView{
		Tag: tag,
		Feed: feedquery.ApplyModeration(ctx, h.prefs, feedquery.ResolveMentionHandles(ctx, h.reader, feed), moderation.UIContextContentList),
	}
}
