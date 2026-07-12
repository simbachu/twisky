package bluesky

import "encoding/json"

const (
	embedImagesType          = "app.bsky.embed.images#view"
	embedGalleryType         = "app.bsky.embed.gallery#view"
	embedRecordType          = "app.bsky.embed.record#view"
	embedRecordWithMediaType = "app.bsky.embed.recordWithMedia#view"
	embedRecordViewType      = "app.bsky.embed.record#viewRecord"
	embedVideoType           = "app.bsky.embed.video#view"
)

type Embed struct {
	Type         string          `json:"$type"`
	Images       []EmbedImage    `json:"images,omitempty"`
	Items        []EmbedImage    `json:"items,omitempty"`
	Record       json.RawMessage `json:"record,omitempty"`
	Media        *Embed          `json:"media,omitempty"`
	CID          string          `json:"cid,omitempty"`
	Playlist     string          `json:"playlist,omitempty"`
	Thumbnail    string          `json:"thumbnail,omitempty"`
	Alt          string          `json:"alt,omitempty"`
	AspectRatio  *AspectRatio    `json:"aspectRatio,omitempty"`
	Presentation string          `json:"presentation,omitempty"`
}

func (e *Embed) MediaImages() []EmbedImage {
	if e == nil {
		return nil
	}
	if e.Type == embedRecordWithMediaType && e.Media != nil {
		return e.Media.MediaImages()
	}
	if len(e.Images) > 0 {
		return e.Images
	}
	return e.Items
}

func (e *Embed) MediaVideos() []*Embed {
	if e == nil {
		return nil
	}
	if e.Type == embedRecordWithMediaType && e.Media != nil {
		return e.Media.MediaVideos()
	}
	if e.Type == embedVideoType && e.Playlist != "" {
		return []*Embed{e}
	}
	return nil
}

func (e *Embed) IsGIF() bool {
	return e != nil && e.Presentation == "gif"
}

// QuotedPost returns the embedded quoted post for record and record-with-media embeds.
func (e *Embed) QuotedPost() *Post {
	if e == nil {
		return nil
	}
	switch e.Type {
	case embedRecordType, embedRecordWithMediaType:
		return postFromRecordEmbed(e.Record)
	default:
		return nil
	}
}

type embedRecordView struct {
	Type   string            `json:"$type"`
	URI    string            `json:"uri"`
	Author Author            `json:"author"`
	Value  PostRecord        `json:"value"`
	Labels []Label           `json:"labels,omitempty"`
	Embeds []json.RawMessage `json:"embeds,omitempty"`
}

func postFromRecordEmbed(raw json.RawMessage) *Post {
	if len(raw) == 0 {
		return nil
	}

	var record embedRecordView
	if err := json.Unmarshal(raw, &record); err != nil || record.URI == "" {
		return nil
	}

	post := Post{
		URI:    record.URI,
		Author: record.Author,
		Record: record.Value,
		Labels: record.Labels,
	}
	for _, nestedRaw := range record.Embeds {
		if embed := parseNestedEmbed(nestedRaw); embed != nil {
			post.Embed = embed
			break
		}
	}
	return &post
}

func parseNestedEmbed(raw json.RawMessage) *Embed {
	if len(raw) == 0 {
		return nil
	}

	var header struct {
		Type string `json:"$type"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil
	}

	switch header.Type {
	case embedImagesType, embedGalleryType, embedRecordType, embedRecordWithMediaType, embedVideoType:
		var embed Embed
		if err := json.Unmarshal(raw, &embed); err != nil {
			return nil
		}
		return &embed
	default:
		return nil
	}
}
