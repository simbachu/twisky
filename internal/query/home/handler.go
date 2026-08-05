package home

import (
	"context"
	"errors"
	"net/http"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/moderation"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/response"
)

type Reader interface {
	GetTimeline(ctx context.Context, req bluesky.TimelineRequest) (*bluesky.AuthorFeedResponse, error)
	GetPosts(ctx context.Context, uris []string) ([]bluesky.Post, error)
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
}

type Handler struct {
	reader Reader
	prefs  moderation.PrefsProvider
}

const HomeFeedLimit = 20

func NewHandler(reader Reader, prefs moderation.PrefsProvider) *Handler {
	if prefs == nil {
		prefs = moderation.DefaultPrefsProvider{}
	}
	return &Handler{reader: reader, prefs: prefs}
}

// HomeView is the read model returned for the home following feed.
type HomeView struct {
	Feed feedquery.FeedView
}

func (HomeView) IsResponse() {}

func (h *Handler) Handle(ctx context.Context, i intent.ViewHome) response.Response {
	items, err := h.reader.GetTimeline(ctx, bluesky.TimelineRequest{
		Limit:  HomeFeedLimit,
		Cursor: i.Cursor,
	})
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "timeline not found"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	feed := feedquery.NewFeedViewFromItems(items.Feed, items.Cursor)
	feed, err = feedquery.EnrichReplyParents(ctx, h.reader, feed)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	return HomeView{
		Feed: feedquery.ApplyModeration(ctx, h.prefs, feedquery.ResolveMentionHandles(ctx, h.reader, feed), moderation.UIContextContentList),
	}
}
