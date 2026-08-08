package home

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/moderation"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/response"
)

type Reader interface {
	GetTimeline(ctx context.Context, req bluesky.TimelineRequest) (*bluesky.AuthorFeedResponse, error)
	GetSavedFeeds(ctx context.Context) ([]bluesky.SavedFeed, error)
	GetFeedGenerators(ctx context.Context, uris []string) ([]bluesky.FeedGenerator, error)
	GetFeed(ctx context.Context, req bluesky.FeedRequest) (*bluesky.AuthorFeedResponse, error)
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

type FeedTab struct {
	Label   string
	Slug    string
	Href    string
	URI     string
	Current bool
}

// HomeView is the read model returned for a home feed.
type HomeView struct {
	Feed  feedquery.FeedView
	Tabs  []FeedTab
	Title string
	Path  string
}

func (HomeView) IsResponse() {}

func (h *Handler) Handle(ctx context.Context, i intent.ViewHome) response.Response {
	tabs, err := h.feedTabs(ctx)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	selected := tabs[0]
	if i.FeedSlug != "" {
		selected = FeedTab{}
		for _, tab := range tabs[1:] {
			if tab.Slug == i.FeedSlug {
				selected = tab
				break
			}
		}
		if selected.URI == "" {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "feed not found"}
		}
	}

	var items *bluesky.AuthorFeedResponse
	if selected.URI == "" {
		items, err = h.reader.GetTimeline(ctx, bluesky.TimelineRequest{
			Limit: HomeFeedLimit, Cursor: i.Cursor,
		})
	} else {
		items, err = h.reader.GetFeed(ctx, bluesky.FeedRequest{
			URI: selected.URI, Limit: HomeFeedLimit, Cursor: i.Cursor,
		})
	}
	if err != nil {
		if errors.Is(err, bluesky.ErrNotFound) {
			return response.ErrorResponse{Status: http.StatusNotFound, Message: "feed not found"}
		}
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	for index := range tabs {
		tabs[index].Current = tabs[index].Slug == selected.Slug
	}
	feed := feedquery.NewFeedViewFromItems(items.Feed, items.Cursor)
	if i.HeadCheck {
		return HomeView{
			Feed:  feedquery.ApplyModeration(ctx, h.prefs, feed, moderation.UIContextContentList),
			Tabs:  tabs,
			Title: selected.Label,
			Path:  selected.Href,
		}
	}
	feed, err = feedquery.EnrichReplyParents(ctx, h.reader, feed)
	if err != nil {
		return response.ErrorResponse{Status: http.StatusBadGateway, Message: "upstream error"}
	}

	return HomeView{
		Feed:  feedquery.ApplyModeration(ctx, h.prefs, feedquery.ResolveMentionHandles(ctx, h.reader, feed), moderation.UIContextContentList),
		Tabs:  tabs,
		Title: selected.Label,
		Path:  selected.Href,
	}
}

func (h *Handler) feedTabs(ctx context.Context) ([]FeedTab, error) {
	saved, err := h.reader.GetSavedFeeds(ctx)
	if err != nil {
		return nil, err
	}
	uris := make([]string, 0, len(saved))
	for _, item := range saved {
		if item.Pinned && item.Type == "feed" && item.URI != "" {
			uris = append(uris, item.URI)
		}
	}
	generators, err := h.reader.GetFeedGenerators(ctx, uris)
	if err != nil {
		return nil, err
	}
	byURI := make(map[string]bluesky.FeedGenerator, len(generators))
	for _, generator := range generators {
		byURI[generator.URI] = generator
	}

	tabs := []FeedTab{{
		Label: "Following",
		Href:  "/",
	}}
	for _, uri := range uris {
		generator, ok := byURI[uri]
		if !ok {
			continue
		}
		slug := feedSlug(generator.DisplayName)
		tabs = append(tabs, FeedTab{
			Label: generator.DisplayName,
			Slug:  slug,
			Href:  "/feed/" + url.PathEscape(slug),
			URI:   uri,
		})
	}
	disambiguateFeedSlugs(tabs)
	for index := range tabs {
		if index == 0 {
			continue
		}
		tabs[index].Href = "/feed/" + url.PathEscape(tabs[index].Slug)
	}
	return tabs, nil
}

func feedSlug(name string) string {
	var builder strings.Builder
	separator := false
	for _, character := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			if separator && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(character)
			separator = false
			continue
		}
		separator = true
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "feed"
	}
	return slug
}

func disambiguateFeedSlugs(tabs []FeedTab) {
	counts := make(map[string]int, len(tabs))
	for _, tab := range tabs[1:] {
		counts[tab.Slug]++
	}
	for index := 1; index < len(tabs); index++ {
		if counts[tabs[index].Slug] < 2 {
			continue
		}
		digest := sha256.Sum256([]byte(tabs[index].URI))
		tabs[index].Slug = fmt.Sprintf("%s-%x", tabs[index].Slug, digest[:4])
	}
}
