package post

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/atproto"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/response"
)

type Reader interface {
	GetProfile(ctx context.Context, actor string) (*bluesky.Profile, error)
	GetPostThread(ctx context.Context, postURI string) (bluesky.ThreadNode, error)
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
}

type Handler struct {
	reader Reader
}

func NewHandler(reader Reader) *Handler {
	return &Handler{reader: reader}
}

func (h *Handler) Handle(ctx context.Context, i intent.ViewPost) response.Response {
	identifier, _, err := actor.ParseSlug(i.Slug)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadRequest, Message: "invalid slug"}
	}

	postID := strings.TrimSpace(i.ID)
	if postID == "" {
		return response.ErrorResponse{Status: http.StatusBadRequest, Message: "invalid post id"}
	}

	profile, err := h.reader.GetProfile(ctx, identifier)
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "actor not found"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	threadNode, err := h.reader.GetPostThread(ctx, atproto.PostURI(profile.DID, postID))
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "post not found"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	root, ok := threadNode.(bluesky.ThreadViewPost)
	if !ok {
		return response.ErrorResponse{Status: http.StatusNotFound, Message: "post not found"}
	}

	view := feedquery.ResolveMentionHandlesInThread(ctx, h.reader, feedquery.NewPostPageView(root))
	return view
}
