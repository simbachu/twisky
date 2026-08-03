package oauth

import (
	"context"
	"fmt"
	"strings"

	"github.com/bluesky-social/indigo/atproto/atclient"
	"github.com/bluesky-social/indigo/atproto/syntax"
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
func (c *SessionClient) CreateLike(ctx context.Context, uri, cid string) error {
	if c == nil || c.client == nil || c.client.AccountDID == nil {
		return fmt.Errorf("oauth: session client not configured")
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
	return c.client.Post(ctx, "com.atproto.repo.createRecord", body, nil)
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

// IsDuplicateLike reports whether err indicates the like already exists.
func IsDuplicateLike(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "record already exists") ||
		strings.Contains(msg, "duplicate")
}
