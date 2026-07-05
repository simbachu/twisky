package feed

import (
	"time"

	"github.com/simbachu/twisky/internal/atproto"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/richtext"
)

type ImageView struct {
	Thumb    string
	Fullsize string
	Alt      string
	Width    int
	Height   int
}

type PostView struct {
	ID                string
	AuthorHandle      string
	AuthorDisplayName string
	AuthorAvatar      string
	Text              string
	TextSegments      []richtext.Segment
	CreatedAt         time.Time
	Images            []ImageView
}

type FeedView struct {
	Posts      []PostView
	NextCursor string
}

func NewFeedView(posts []bluesky.Post, cursor string) FeedView {
	views := make([]PostView, 0, len(posts))
	for _, post := range posts {
		views = append(views, NewPostView(post))
	}
	return FeedView{
		Posts:      views,
		NextCursor: cursor,
	}
}

func NewPostView(post bluesky.Post) PostView {
	id, _ := atproto.PostRkey(post.URI)
	view := PostView{
		ID:                id,
		AuthorHandle:      post.Author.Handle,
		AuthorDisplayName: post.Author.DisplayName,
		AuthorAvatar:      post.Author.Avatar,
		Text:              post.Record.Text,
		TextSegments:      richtext.BuildSegments(post.Record.Text, post.Record.Facets),
		CreatedAt:         post.Record.CreatedAt,
	}
	if post.Embed != nil {
		for _, image := range post.Embed.MediaImages() {
			imageView := ImageView{
				Thumb:    image.ThumbURL(),
				Fullsize: image.Fullsize,
				Alt:      image.Alt,
			}
			if image.AspectRatio != nil {
				imageView.Width = image.AspectRatio.Width
				imageView.Height = image.AspectRatio.Height
			}
			view.Images = append(view.Images, imageView)
		}
	}
	return view
}
