package bluesky

import (
	"encoding/json"
	"time"
)

// ParseFeedResponse decodes a getTimeline/getAuthorFeed/getFeed JSON body, skipping malformed feed items.
func ParseFeedResponse(body []byte) ([]FeedItem, string, error) {
	var raw struct {
		Feed   []json.RawMessage `json:"feed"`
		Cursor string            `json:"cursor,omitempty"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, "", err
	}
	return ParseFeedItems(raw.Feed), raw.Cursor, nil
}

// ParseFeedItems decodes feed view posts one at a time, omitting items that fail to parse.
func ParseFeedItems(rawItems []json.RawMessage) []FeedItem {
	items := make([]FeedItem, 0, len(rawItems))
	for _, raw := range rawItems {
		var item FeedItem
		if err := json.Unmarshal(raw, &item); err != nil {
			continue
		}
		if item.Post.URI == "" {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (r *PostRecord) UnmarshalJSON(data []byte) error {
	var raw struct {
		Text       string          `json:"text"`
		CreatedAt  json.RawMessage `json:"createdAt"`
		Facets     []Facet         `json:"facets,omitempty"`
		Reply      *RecordReplyRef `json:"reply,omitempty"`
		SelfLabels *SelfLabels     `json:"labels,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.Text = raw.Text
	r.Facets = raw.Facets
	r.Reply = raw.Reply
	r.SelfLabels = raw.SelfLabels
	r.CreatedAt = parseCreatedAt(raw.CreatedAt)
	return nil
}

// FeedListResponse is the feed+cursor body returned by timeline and author-feed XRPC methods.
type FeedListResponse struct {
	Feed   []FeedItem
	Cursor string
}

func (r *FeedListResponse) UnmarshalJSON(data []byte) error {
	feed, cursor, err := ParseFeedResponse(data)
	if err != nil {
		return err
	}
	r.Feed = feed
	r.Cursor = cursor
	return nil
}

func parseCreatedAt(raw json.RawMessage) time.Time {
	if len(raw) == 0 || string(raw) == "null" {
		return time.Time{}
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return time.Time{}
	}
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}
