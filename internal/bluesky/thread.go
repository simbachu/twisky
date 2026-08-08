package bluesky

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

const (
	threadViewPostType = "app.bsky.feed.defs#threadViewPost"
	notFoundPostType   = "app.bsky.feed.defs#notFoundPost"
	blockedPostType    = "app.bsky.feed.defs#blockedPost"

	// Default caps for getPostThread so viral threads cannot return unbounded trees.
	DefaultThreadDepth        = 6
	DefaultThreadParentHeight = 25
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

// PostThreadOpts bounds how much of the reply and ancestor trees AppView returns.
// Zero Depth or ParentHeight means use the corresponding DefaultThread* constant.
type PostThreadOpts struct {
	Depth        int
	ParentHeight int
}

type getPostThreadResponse struct {
	Thread json.RawMessage `json:"thread"`
}

func (c *Client) GetPostThread(ctx context.Context, postURI string, opts *PostThreadOpts) (ThreadNode, error) {
	endpoint, err := url.Parse(c.baseURL + "/app.bsky.feed.getPostThread")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("uri", postURI)
	query.Set("depth", strconv.Itoa(threadOptOrDefault(opts, true)))
	query.Set("parentHeight", strconv.Itoa(threadOptOrDefault(opts, false)))
	endpoint.RawQuery = query.Encode()

	var threadResp getPostThreadResponse
	if err := c.doGet(ctx, endpoint.String(), &threadResp); err != nil {
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

func threadOptOrDefault(opts *PostThreadOpts, depth bool) int {
	if opts != nil {
		if depth && opts.Depth > 0 {
			return opts.Depth
		}
		if !depth && opts.ParentHeight > 0 {
			return opts.ParentHeight
		}
	}
	if depth {
		return DefaultThreadDepth
	}
	return DefaultThreadParentHeight
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
