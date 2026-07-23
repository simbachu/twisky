package post

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/atproto"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/moderation"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/response"
)

// countsCacheTTL bounds how long a counts-only fetch is reused across
// concurrent pollers of the same post.
const countsCacheTTL = 5 * time.Second

type Reader interface {
	GetProfile(ctx context.Context, actor string) (*bluesky.Profile, error)
	GetPostThread(ctx context.Context, postURI string) (bluesky.ThreadNode, error)
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
	GetPosts(ctx context.Context, uris []string) ([]bluesky.Post, error)
}

type Handler struct {
	reader      Reader
	prefs       moderation.PrefsProvider
	countsCache *countsCache
}

func NewHandler(reader Reader, prefs moderation.PrefsProvider) *Handler {
	if prefs == nil {
		prefs = moderation.DefaultPrefsProvider{}
	}
	return &Handler{reader: reader, prefs: prefs, countsCache: newCountsCache(countsCacheTTL)}
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

	if i.Part == feedquery.PostPagePartCounts {
		return h.handleCounts(ctx, i.Slug, postID, profile.DID)
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

	view := feedquery.NewPostPageView(root, i.Part)
	view = feedquery.ResolveMentionHandlesInThread(ctx, h.reader, view)
	view = feedquery.ApplyModerationToPostPage(ctx, h.prefs, view)
	return view
}

// handleCounts serves the cheap counts-only fragment via GetPosts instead of
// the heavier GetPostThread, coalescing concurrent requests for the same post
// through countsCache.
func (h *Handler) handleCounts(ctx context.Context, slug, postID, did string) response.Response {
	uri := atproto.PostURI(did, postID)
	key := slug + "/" + postID

	bskyPost, err := h.countsCache.Get(ctx, key, func(ctx context.Context) (bluesky.Post, error) {
		return h.fetchPostForCounts(ctx, uri)
	})
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "post not found"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	return feedquery.PostPageView{Post: feedquery.NewPostView(bskyPost)}
}

func (h *Handler) fetchPostForCounts(ctx context.Context, uri string) (bluesky.Post, error) {
	posts, err := h.reader.GetPosts(ctx, []string{uri})
	if err != nil {
		return bluesky.Post{}, err
	}
	if len(posts) == 0 {
		return bluesky.Post{}, bluesky.ErrNotFound
	}
	return posts[0], nil
}
