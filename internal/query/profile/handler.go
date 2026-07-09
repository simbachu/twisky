package profile

import (
	"context"
	"errors"
	"net/http"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/response"
)

type Reader interface {
	GetProfile(ctx context.Context, actor string) (*bluesky.Profile, error)
	GetAuthorFeed(ctx context.Context, req bluesky.AuthorFeedRequest) (*bluesky.AuthorFeedResponse, error)
	GetPosts(ctx context.Context, uris []string) ([]bluesky.Post, error)
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
}

type Handler struct {
	reader Reader
}

const ProfileFeedLimit = 20

func NewHandler(reader Reader) *Handler {
	return &Handler{reader: reader}
}

type Tab string

const (
	TabPosts Tab = "posts"
	TabMedia Tab = "media"
)

// ProfileView is the read model returned for a profile page.
type ProfileView struct {
	DID         string // No need to surface this to the user
	Handle      string // format @handle.url
	DisplayName string
	Description string
	Avatar      string // url
	Followers   int
	Following   int
	Posts       int
	Tab         Tab
	Feed        feedquery.FeedView
}

func (ProfileView) IsResponse() {}

func (h *Handler) Handle(ctx context.Context, i intent.ViewProfile) response.Response {
	identifier, _, err := actor.ParseSlug(i.Slug)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadRequest, Message: "invalid slug"}
	}

	profile, err := h.reader.GetProfile(ctx, identifier)
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "actor not found"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	filter := bluesky.FilterPostsNoReplies
	tab := TabPosts
	if i.Tab == intent.ProfileTabMedia {
		filter = bluesky.FilterPostsWithMedia
		tab = TabMedia
	}

	items, err := h.reader.GetAuthorFeed(ctx, bluesky.AuthorFeedRequest{
		Actor:  identifier,
		Filter: filter,
		Limit:  ProfileFeedLimit,
		Cursor: i.Cursor,
	})
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "actor not found"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	feed := feedquery.NewFeedViewFromItems(items.Feed, items.Cursor)
	feed, err = feedquery.EnrichReplyParents(ctx, h.reader, feed)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	return ProfileView{
		DID:         profile.DID,
		Handle:      profile.Handle,
		DisplayName: actor.Name(profile.DisplayName, profile.Handle),
		Description: profile.Description,
		Avatar:      profile.Avatar,
		Followers:   profile.Followers,
		Following:   profile.Following,
		Posts:       profile.Posts,
		Tab:         tab,
		Feed:        feedquery.ResolveMentionHandles(ctx, h.reader, feed),
	}
}
