package profile

import (
	"context"
	"errors"
	"net/http"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/identity"
	"github.com/simbachu/twisky/internal/intent"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/moderation"
	"github.com/simbachu/twisky/internal/response"
	"github.com/simbachu/twisky/internal/richtext"
)

type Reader interface {
	GetProfile(ctx context.Context, actor string) (*bluesky.Profile, error)
	GetAuthorFeed(ctx context.Context, req bluesky.AuthorFeedRequest) (*bluesky.AuthorFeedResponse, error)
	GetPosts(ctx context.Context, uris []string) ([]bluesky.Post, error)
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
}

type Handler struct {
	reader Reader
	prefs  moderation.PrefsProvider
	dir    *identity.Directory
}

const ProfileFeedLimit = 20

func NewHandler(reader Reader, prefs moderation.PrefsProvider, dir *identity.Directory) *Handler {
	if prefs == nil {
		prefs = moderation.DefaultPrefsProvider{}
	}
	return &Handler{reader: reader, prefs: prefs, dir: dir}
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
	Description         string
	DescriptionSegments []richtext.Segment
	Avatar              string // url
	Banner              string // url
	Followers   int
	Following   int
	Posts       int
	Tab         Tab
	PinnedPostMaybe *feedquery.PostView
	Feed        feedquery.FeedView
	Labels           []moderation.ProfileLabelView
	AvatarModeration feedquery.ModerationView
	IsLabeler        bool
}

func (ProfileView) IsResponse() {}

func (h *Handler) Handle(ctx context.Context, i intent.ViewProfile) response.Response {
	slug, err := actor.ParseSlug(i.Slug)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid handle or DID"}
	}
	referent := slug.Referent()

	did, errResp := h.resolveDID(ctx, slug)
	if errResp != nil {
		return *errResp
	}

	profile, err := h.reader.GetProfile(ctx, did)
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "Could not find " + referent}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "Failed to resolve " + referent}
	}
	if h.dir != nil {
		h.dir.ObserveProfile(profile)
	}
	displayHandle := profile.Handle
	if h.dir != nil {
		displayHandle = h.dir.DisplayHandle(profile.Handle, profile.DID)
	}

	filter := bluesky.FilterPostsNoReplies
	tab := TabPosts
	if i.Tab == intent.ProfileTabMedia {
		filter = bluesky.FilterPostsWithMedia
		tab = TabMedia
	}

	items, err := h.reader.GetAuthorFeed(ctx, bluesky.AuthorFeedRequest{
		Actor:  did,
		Filter: filter,
		Limit:  ProfileFeedLimit,
		Cursor: i.Cursor,
	})
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "Could not find posts for " + referent}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "Failed to load posts for " + referent}
	}

	feed := feedquery.NewFeedViewFromItems(items.Feed, items.Cursor)
	if h.dir != nil {
		feedquery.ObserveAuthors(h.dir, feed)
	}
	if i.HeadCheck {
		return ProfileView{
			DID:    profile.DID,
			Handle: displayHandle,
			Tab:    tab,
			Feed:   feedquery.ApplyModeration(ctx, h.prefs, feed, moderation.UIContextContentList),
		}
	}
	feed, err = feedquery.EnrichReplyParents(ctx, h.reader, feed)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "Failed to load reply context"}
	}

	var descriptionSegments []richtext.Segment
	if len(profile.DescriptionFacets) > 0 {
		descriptionSegments = richtext.BuildSegments(profile.Description, profile.DescriptionFacets)
	} else {
		descriptionSegments = richtext.DetectSegments(profile.Description)
	}
	descriptionSegments = feedquery.ResolveMentionHandlesInSegments(ctx, h.reader, descriptionSegments)

	moderatedFeed := feedquery.ApplyModeration(ctx, h.prefs, feedquery.ResolveMentionHandles(ctx, h.reader, feedquery.ApplyIdentityFeed(h.dir, feed)), moderation.UIContextContentList)

	prefs := h.prefs.Prefs(ctx)
	profileLabels := moderation.LabelsFromBluesky(profile.Labels)
	avatarUI := moderation.EvaluateProfileAvatar(ctx, h.prefs, profile.DID, profileLabels)
	displayLabels := moderation.ProfileLabelsForDisplay(profileLabels, profile.DID, prefs)
	displayLabels, err = h.enrichProfileLabels(ctx, displayLabels, profile)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "Failed to load moderation label providers"}
	}

	isLabeler := actor.IsLabelerAccount(displayHandle, profile.DID, profile.Associated != nil && profile.Associated.Labeler)

	return ProfileView{
		DID:                 profile.DID,
		Handle:              displayHandle,
		DisplayName:         actor.Name(profile.DisplayName, displayHandle),
		Description:         profile.Description,
		DescriptionSegments: descriptionSegments,
		Avatar:              profile.Avatar,
		Banner:              profile.Banner,
		Followers:           profile.Followers,
		Following:           profile.Following,
		Posts:               profile.Posts,
		Tab:                 tab,
		PinnedPostMaybe:     h.pinnedPostMaybe(ctx, profile, i.Cursor),
		Feed:                moderatedFeed,
		Labels:              displayLabels,
		AvatarModeration: feedquery.ModerationView{
			BlurAvatar: avatarUI.BlurAvatar,
			AvatarText: avatarUI.PrimaryMessage(),
		},
		IsLabeler: isLabeler,
	}
}

func (h *Handler) resolveDID(ctx context.Context, slug actor.Slug) (string, *response.ErrorResponse) {
	if slug.Kind == actor.KindDID {
		return slug.Identifier, nil
	}
	if h.dir == nil {
		referent := slug.Referent()
		profile, err := h.reader.GetProfile(ctx, slug.Identifier)
		if err != nil {
			if errors.Is(err, bluesky.ErrNotFound) {
				return "", &response.ErrorResponse{Status: http.StatusNotFound, Message: "Could not find " + referent}
			}
			return "", &response.ErrorResponse{Status: http.StatusBadGateway, Message: "Failed to resolve " + referent}
		}
		if profile == nil || profile.DID == "" {
			return "", &response.ErrorResponse{Status: http.StatusBadGateway, Message: "Failed to resolve " + referent}
		}
		return profile.DID, nil
	}

	referent := slug.Referent()
	did, err := h.dir.Resolve(ctx, slug, h.reader)
	if err != nil {
		if errors.Is(err, identity.ErrNotFound) || errors.Is(err, bluesky.ErrNotFound) {
			return "", &response.ErrorResponse{Status: http.StatusNotFound, Message: "Could not find " + referent}
		}
		return "", &response.ErrorResponse{Status: http.StatusBadGateway, Message: "Failed to resolve " + referent}
	}
	if did == "" {
		return "", &response.ErrorResponse{Status: http.StatusBadGateway, Message: "Failed to resolve " + referent}
	}
	return did, nil
}

func (h *Handler) pinnedPostMaybe(ctx context.Context, profile *bluesky.Profile, cursor string) *feedquery.PostView {
	if profile.PinnedPost == nil || cursor != "" {
		return nil
	}

	posts, err := h.reader.GetPosts(ctx, []string{profile.PinnedPost.URI})
	if err != nil || len(posts) == 0 {
		return nil
	}

	pinned := feedquery.InsetPostView(feedquery.ApplyIdentity(h.dir, feedquery.NewPostView(posts[0])))
	moderated := feedquery.ApplyModeration(ctx, h.prefs, feedquery.ResolveMentionHandles(ctx, h.reader, feedquery.FeedView{
		Posts: []feedquery.PostView{pinned},
	}), moderation.UIContextContentList)
	if len(moderated.Posts) == 0 {
		return nil
	}
	return &moderated.Posts[0]
}

func (h *Handler) enrichProfileLabels(ctx context.Context, labels []moderation.ProfileLabelView, profile *bluesky.Profile) ([]moderation.ProfileLabelView, error) {
	if len(labels) == 0 {
		return labels, nil
	}

	labelerDIDs := make([]string, 0)
	seen := make(map[string]struct{})
	for _, label := range labels {
		if label.LabelerDID == profile.DID {
			continue
		}
		if _, ok := seen[label.LabelerDID]; ok {
			continue
		}
		seen[label.LabelerDID] = struct{}{}
		labelerDIDs = append(labelerDIDs, label.LabelerDID)
	}

	profilesByDID := make(map[string]bluesky.Profile)
	if len(labelerDIDs) > 0 {
		profiles, err := h.reader.GetProfiles(ctx, labelerDIDs)
		if err != nil {
			return nil, err
		}
		for _, labelerProfile := range profiles {
			profilesByDID[labelerProfile.DID] = labelerProfile
		}
	}

	enriched := make([]moderation.ProfileLabelView, len(labels))
	for i, label := range labels {
		enriched[i] = label
		if label.LabelerDID == profile.DID {
			enriched[i].LabelerHandle = profile.Handle
			enriched[i].LabelerAvatar = profile.Avatar
			enriched[i].LabelerIsLabeler = actor.IsLabelerAccount(profile.Handle, profile.DID, profile.Associated != nil && profile.Associated.Labeler)
			continue
		}
		if labelerProfile, ok := profilesByDID[label.LabelerDID]; ok {
			enriched[i].LabelerHandle = labelerProfile.Handle
			enriched[i].LabelerAvatar = labelerProfile.Avatar
			enriched[i].LabelerIsLabeler = actor.IsLabelerAccount(
				labelerProfile.Handle,
				labelerProfile.DID,
				labelerProfile.Associated != nil && labelerProfile.Associated.Labeler,
			)
		} else {
			enriched[i].LabelerHandle = moderation.LabelerProfileSlug(label.LabelerDID)
			enriched[i].LabelerIsLabeler = actor.IsLabelerAccount("", label.LabelerDID, false)
		}
	}
	return enriched, nil
}
