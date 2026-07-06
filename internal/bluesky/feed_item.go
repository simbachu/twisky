package bluesky

import "encoding/json"

const (
	postViewType     = "app.bsky.feed.defs#postView"
	feedNotFoundType = "app.bsky.feed.defs#notFoundPost"
	feedBlockedType  = "app.bsky.feed.defs#blockedPost"
)

type FeedItem struct {
	Post  Post          `json:"post"`
	Reply *ReplyContext `json:"reply,omitempty"`
}

type ReplyContext struct {
	Parent *Post
}

func (r *ReplyContext) UnmarshalJSON(data []byte) error {
	var raw struct {
		Parent json.RawMessage `json:"parent"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parent, err := parseFeedPostUnion(raw.Parent)
	if err != nil {
		return err
	}
	r.Parent = parent
	return nil
}

func parseFeedPostUnion(raw json.RawMessage) (*Post, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var probe struct {
		Type     string `json:"$type"`
		NotFound bool   `json:"notFound"`
		Blocked  bool   `json:"blocked"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, err
	}
	if probe.NotFound || probe.Blocked || probe.Type == feedNotFoundType || probe.Type == feedBlockedType {
		return nil, nil
	}

	var post Post
	if err := json.Unmarshal(raw, &post); err != nil {
		return nil, err
	}
	if post.URI == "" {
		return nil, nil
	}
	return &post, nil
}
