package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bluesky-social/indigo/atproto/atclient"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/simbachu/twisky/internal/atproto"
	"github.com/simbachu/twisky/internal/bluesky"
)

const appViewService = "did:web:api.bsky.app#bsky_appview"

// SessionClient performs authenticated PDS writes and AppView reads for a resumed OAuth session.
type SessionClient struct {
	client *atclient.APIClient
}

func NewSessionClient(client *atclient.APIClient) *SessionClient {
	return &SessionClient{client: client}
}

// CreateLike writes an app.bsky.feed.like record for the subject strong-ref.
func (c *SessionClient) CreateLike(ctx context.Context, uri, cid string) (string, error) {
	if c == nil || c.client == nil || c.client.AccountDID == nil {
		return "", fmt.Errorf("oauth: session client not configured")
	}
	body := map[string]any{
		"repo":       c.client.AccountDID.String(),
		"collection": "app.bsky.feed.like",
		"record": map[string]any{
			"$type": "app.bsky.feed.like",
			"subject": map[string]any{
				"uri": uri,
				"cid": cid,
			},
			"createdAt": syntax.DatetimeNow(),
		},
	}
	var resp struct {
		URI string `json:"uri"`
	}
	if err := c.client.Post(ctx, "com.atproto.repo.createRecord", body, &resp); err != nil {
		return "", err
	}
	return resp.URI, nil
}

// DeleteLike removes an app.bsky.feed.like record by AT URI.
func (c *SessionClient) DeleteLike(ctx context.Context, recordURI string) error {
	return c.deleteRecord(ctx, recordURI)
}

// CreateRepost writes an app.bsky.feed.repost record for the subject strong-ref.
func (c *SessionClient) CreateRepost(ctx context.Context, uri, cid string) (string, error) {
	if c == nil || c.client == nil || c.client.AccountDID == nil {
		return "", fmt.Errorf("oauth: session client not configured")
	}
	body := map[string]any{
		"repo":       c.client.AccountDID.String(),
		"collection": "app.bsky.feed.repost",
		"record": map[string]any{
			"$type": "app.bsky.feed.repost",
			"subject": map[string]any{
				"uri": uri,
				"cid": cid,
			},
			"createdAt": syntax.DatetimeNow(),
		},
	}
	var resp struct {
		URI string `json:"uri"`
	}
	if err := c.client.Post(ctx, "com.atproto.repo.createRecord", body, &resp); err != nil {
		return "", err
	}
	return resp.URI, nil
}

// DeleteRepost removes an app.bsky.feed.repost record by AT URI.
func (c *SessionClient) DeleteRepost(ctx context.Context, recordURI string) error {
	return c.deleteRecord(ctx, recordURI)
}

func (c *SessionClient) deleteRecord(ctx context.Context, recordURI string) error {
	if c == nil || c.client == nil || c.client.AccountDID == nil {
		return fmt.Errorf("oauth: session client not configured")
	}
	parsed, err := atproto.ParseRecordURI(recordURI)
	if err != nil {
		return err
	}
	body := map[string]any{
		"repo":       parsed.AuthorDID(),
		"collection": parsed.Collection(),
		"rkey":       parsed.Rkey(),
	}
	return c.client.Post(ctx, "com.atproto.repo.deleteRecord", body, nil)
}

// GetPosts fetches posts via the AppView proxy (includes viewer state when authenticated).
func (c *SessionClient) GetPosts(ctx context.Context, uris []string) ([]bluesky.Post, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("oauth: session client not configured")
	}
	if len(uris) == 0 {
		return nil, nil
	}
	appview := c.client.WithService(appViewService)
	const chunk = 25
	out := make([]bluesky.Post, 0, len(uris))
	for start := 0; start < len(uris); start += chunk {
		end := start + chunk
		if end > len(uris) {
			end = len(uris)
		}
		var resp struct {
			Posts []bluesky.Post `json:"posts"`
		}
		params := map[string]any{"uris": uris[start:end]}
		if err := appview.Get(ctx, "app.bsky.feed.getPosts", params, &resp); err != nil {
			return nil, err
		}
		out = append(out, resp.Posts...)
	}
	return out, nil
}

// GetTimeline fetches the authenticated user's following feed via the AppView proxy.
func (c *SessionClient) GetTimeline(ctx context.Context, req bluesky.TimelineRequest) (*bluesky.AuthorFeedResponse, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("oauth: session client not configured")
	}
	appview := c.client.WithService(appViewService)
	params := map[string]any{}
	if req.Limit > 0 {
		params["limit"] = req.Limit
	}
	if req.Cursor != "" {
		params["cursor"] = req.Cursor
	}
	var resp struct {
		Feed   []bluesky.FeedItem `json:"feed"`
		Cursor string             `json:"cursor,omitempty"`
	}
	if err := appview.Get(ctx, "app.bsky.feed.getTimeline", params, &resp); err != nil {
		return nil, err
	}
	return &bluesky.AuthorFeedResponse{
		Feed:   resp.Feed,
		Cursor: resp.Cursor,
	}, nil
}

// GetSavedFeeds returns the authenticated user's saved-feed preferences.
func (c *SessionClient) GetSavedFeeds(ctx context.Context) ([]bluesky.SavedFeed, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("oauth: session client not configured")
	}
	appview := c.client.WithService(appViewService)
	var resp struct {
		Preferences []json.RawMessage `json:"preferences"`
	}
	if err := appview.Get(ctx, "app.bsky.actor.getPreferences", map[string]any{}, &resp); err != nil {
		return nil, err
	}
	return parseSavedFeeds(resp.Preferences)
}

func parseSavedFeeds(preferences []json.RawMessage) ([]bluesky.SavedFeed, error) {
	var legacy []bluesky.SavedFeed
	for _, raw := range preferences {
		var header struct {
			Type string `json:"$type"`
		}
		if err := json.Unmarshal(raw, &header); err != nil {
			return nil, err
		}
		switch header.Type {
		case "app.bsky.actor.defs#savedFeedsPrefV2":
			var pref struct {
				Items []bluesky.SavedFeed `json:"items"`
			}
			if err := json.Unmarshal(raw, &pref); err != nil {
				return nil, err
			}
			return pref.Items, nil
		case "app.bsky.actor.defs#savedFeedsPref":
			var pref struct {
				Pinned []string `json:"pinned"`
			}
			if err := json.Unmarshal(raw, &pref); err != nil {
				return nil, err
			}
			legacy = make([]bluesky.SavedFeed, 0, len(pref.Pinned))
			for _, uri := range pref.Pinned {
				legacy = append(legacy, bluesky.SavedFeed{
					ID:     uri,
					Pinned: true,
					Type:   "feed",
					URI:    uri,
				})
			}
		}
	}
	return legacy, nil
}

// GetFeedGenerators resolves feed-generator metadata via the AppView proxy.
func (c *SessionClient) GetFeedGenerators(ctx context.Context, uris []string) ([]bluesky.FeedGenerator, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("oauth: session client not configured")
	}
	if len(uris) == 0 {
		return nil, nil
	}
	appview := c.client.WithService(appViewService)
	const chunk = 25
	out := make([]bluesky.FeedGenerator, 0, len(uris))
	for start := 0; start < len(uris); start += chunk {
		end := min(start+chunk, len(uris))
		var resp struct {
			Feeds []bluesky.FeedGenerator `json:"feeds"`
		}
		if err := appview.Get(ctx, "app.bsky.feed.getFeedGenerators", map[string]any{
			"feeds": uris[start:end],
		}, &resp); err != nil {
			return nil, err
		}
		out = append(out, resp.Feeds...)
	}
	return out, nil
}

// GetFeed fetches an authenticated custom feed via the AppView proxy.
func (c *SessionClient) GetFeed(ctx context.Context, req bluesky.FeedRequest) (*bluesky.AuthorFeedResponse, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("oauth: session client not configured")
	}
	appview := c.client.WithService(appViewService)
	params := map[string]any{"feed": req.URI}
	if req.Limit > 0 {
		params["limit"] = req.Limit
	}
	if req.Cursor != "" {
		params["cursor"] = req.Cursor
	}
	var resp struct {
		Feed   []bluesky.FeedItem `json:"feed"`
		Cursor string             `json:"cursor,omitempty"`
	}
	if err := appview.Get(ctx, "app.bsky.feed.getFeed", params, &resp); err != nil {
		return nil, err
	}
	return &bluesky.AuthorFeedResponse{Feed: resp.Feed, Cursor: resp.Cursor}, nil
}

// GetProfiles fetches actor profiles via the AppView proxy.
func (c *SessionClient) GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("oauth: session client not configured")
	}
	if len(actors) == 0 {
		return nil, nil
	}
	appview := c.client.WithService(appViewService)
	var resp struct {
		Profiles []bluesky.Profile `json:"profiles"`
	}
	params := map[string]any{"actors": actors}
	if err := appview.Get(ctx, "app.bsky.actor.getProfiles", params, &resp); err != nil {
		return nil, err
	}
	return resp.Profiles, nil
}

// IsDuplicateRecord reports whether err indicates the record already exists.
func IsDuplicateRecord(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "record already exists") ||
		strings.Contains(msg, "duplicate")
}

// IsDuplicateLike reports whether err indicates the like already exists.
func IsDuplicateLike(err error) bool {
	return IsDuplicateRecord(err)
}

// IsRecordNotFound reports whether err indicates the record does not exist.
func IsRecordNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "could not locate record") ||
		strings.Contains(msg, "recordnotfound")
}
