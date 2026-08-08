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

// threadCacheTTL bounds how long a GetPostThread fetch is reused across
// full page, ancestors, and replies requests for the same post.
const threadCacheTTL = 20 * time.Second

// threadDepth / threadParentHeight cap upstream getPostThread payload size.
const (
	threadDepth        = 6
	threadParentHeight = 25
)

type Reader interface {
	GetProfile(ctx context.Context, actor string) (*bluesky.Profile, error)
	GetPostThread(ctx context.Context, postURI string, opts *bluesky.PostThreadOpts) (bluesky.ThreadNode, error)
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
	GetPosts(ctx context.Context, uris []string) ([]bluesky.Post, error)
}

type Handler struct {
	reader      Reader
	prefs       moderation.PrefsProvider
	countsCache *ttlCache[bluesky.Post]
	threadCache *ttlCache[bluesky.ThreadNode]
}

func NewHandler(reader Reader, prefs moderation.PrefsProvider) *Handler {
	if prefs == nil {
		prefs = moderation.DefaultPrefsProvider{}
	}
	return &Handler{
		reader:      reader,
		prefs:       prefs,
		countsCache: newCountsCache(countsCacheTTL),
		threadCache: newThreadCache(threadCacheTTL),
	}
}

func (h *Handler) Handle(ctx context.Context, i intent.ViewPost) response.Response {
	slug, err := actor.ParseSlug(i.Slug)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadRequest, Message: "invalid slug"}
	}

	postID := strings.TrimSpace(i.ID)
	if postID == "" {
		return response.ErrorResponse{Status: http.StatusBadRequest, Message: "invalid post id"}
	}

	did, errResp := h.resolveDID(ctx, slug)
	if errResp != nil {
		return *errResp
	}

	if i.Part == feedquery.PostPagePartCounts {
		return h.handleCounts(ctx, i.Slug, postID, did)
	}

	threadNode, err := h.getThread(ctx, i.Slug, postID, did)
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "post not found"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	return h.postPageFromThread(ctx, threadNode, i.Part)
}

// resolveDID returns the account DID for the post URI. When the slug is already
// a DID, GetProfile is skipped entirely (counts/replies/full page only need DID).
func (h *Handler) resolveDID(ctx context.Context, slug actor.Slug) (string, *response.ErrorResponse) {
	if slug.Kind == actor.KindDID {
		return slug.Identifier, nil
	}

	profile, err := h.reader.GetProfile(ctx, slug.Identifier)
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return "", &response.ErrorResponse{Status: http.StatusNotFound, Message: "actor not found"}
		}
		return "", &response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}
	if profile == nil || profile.DID == "" {
		return "", &response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}
	return profile.DID, nil
}

// handleCounts serves the cheap counts-only fragment via GetPosts instead of
// the heavier GetPostThread, coalescing concurrent requests for the same post
// through countsCache.
func (h *Handler) handleCounts(ctx context.Context, slug, postID, did string) response.Response {
	uri := atproto.NewPostURI(did, postID).String()
	key := slug + "/" + postID

	bskyPost, err := h.countsCache.Get(ctx, key, func(ctx context.Context) (bluesky.Post, error) {
		return h.fetchPostForCounts(ctx, uri)
	})
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "post not found"}
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	return feedquery.PostPageView{Post: feedquery.NewPostView(bskyPost)}
}

func (h *Handler) getThread(ctx context.Context, slug, postID, did string) (bluesky.ThreadNode, error) {
	uri := atproto.NewPostURI(did, postID).String()
	key := slug + "/" + postID
	opts := &bluesky.PostThreadOpts{Depth: threadDepth, ParentHeight: threadParentHeight}
	return h.threadCache.Get(ctx, key, func(ctx context.Context) (bluesky.ThreadNode, error) {
		return h.reader.GetPostThread(ctx, uri, opts)
	})
}

func (h *Handler) postPageFromThread(ctx context.Context, threadNode bluesky.ThreadNode, part string) response.Response {
	root, ok := threadNode.(bluesky.ThreadViewPost)
	if !ok {
		return response.ErrorResponse{Status: http.StatusNotFound, Message: "post not found"}
	}

	view := feedquery.NewPostPageView(root, part)
	view = feedquery.ResolveMentionHandlesInThread(ctx, h.reader, view)
	view = feedquery.ApplyModerationToPostPage(ctx, h.prefs, view)
	return view
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
