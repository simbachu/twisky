package bluesky

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	threadViewPostType = "app.bsky.feed.defs#threadViewPost"
	notFoundPostType   = "app.bsky.feed.defs#notFoundPost"
	blockedPostType    = "app.bsky.feed.defs#blockedPost"
)

type ThreadNode interface {
	threadNode()
}

type ThreadViewPost struct {
	Post    Post
	Parent  ThreadNode
	Replies []ThreadNode
}

func (ThreadViewPost) threadNode() {}

type NotFoundPost struct {
	URI string
}

func (NotFoundPost) threadNode() {}

type BlockedPost struct {
	URI string
}

func (BlockedPost) threadNode() {}

type getPostThreadResponse struct {
	Thread json.RawMessage `json:"thread"`
}

func (c *Client) GetPostThread(ctx context.Context, postURI string) (ThreadNode, error) {
	endpoint, err := url.Parse(c.baseURL + "/app.bsky.feed.getPostThread")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("uri", postURI)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("bluesky api: %s", apiErr.Message)
		}
		return nil, fmt.Errorf("bluesky api: status %d", resp.StatusCode)
	}

	var threadResp getPostThreadResponse
	if err := json.Unmarshal(body, &threadResp); err != nil {
		return nil, err
	}

	node, err := parseThreadNode(threadResp.Thread)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, ErrNotFound
	}
	return node, nil
}

func parseThreadNode(raw json.RawMessage) (ThreadNode, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var header struct {
		Type string `json:"$type"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, err
	}

	switch header.Type {
	case threadViewPostType:
		var payload struct {
			Post    Post              `json:"post"`
			Parent  json.RawMessage   `json:"parent,omitempty"`
			Replies []json.RawMessage `json:"replies,omitempty"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, err
		}

		parent, err := parseThreadNode(payload.Parent)
		if err != nil {
			return nil, err
		}

		replies := make([]ThreadNode, 0, len(payload.Replies))
		for _, replyRaw := range payload.Replies {
			reply, err := parseThreadNode(replyRaw)
			if err != nil {
				return nil, err
			}
			if reply != nil {
				replies = append(replies, reply)
			}
		}

		return ThreadViewPost{
			Post:    payload.Post,
			Parent:  parent,
			Replies: replies,
		}, nil
	case notFoundPostType:
		var payload struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, err
		}
		return NotFoundPost{URI: payload.URI}, nil
	case blockedPostType:
		var payload struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, err
		}
		return BlockedPost{URI: payload.URI}, nil
	default:
		return nil, fmt.Errorf("bluesky api: unknown thread node type %q", header.Type)
	}
}
