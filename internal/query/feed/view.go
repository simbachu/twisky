package feed

import (
	"time"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/atproto"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/moderation"
	"github.com/simbachu/twisky/internal/richtext"
)

type ModerationView struct {
	Filtered   bool
	Blurred    bool
	NoOverride bool
	AlertText  string
	BlurMedia  bool
}

type ImageView struct {
	Thumb    string
	Fullsize string
	Alt      string
	Width    int
	Height   int
}

type AuthorView struct {
	Handle      string
	DisplayName string
	Avatar      string
}

type PostView struct {
	ID                string
	AuthorHandle      string
	AuthorDisplayName string
	AuthorAvatar      string
	LikeCount         int
	RepostCount       int
	ReplyCount        int
	Text              string
	TextSegments      []richtext.Segment
	CreatedAt         time.Time
	Images            []ImageView
	RepostedByMaybe   *AuthorView
	ReplyParentMaybe  *PostView
	QuotedPostMaybe   *PostView
	Moderation        ModerationView
	replyParentURI    string
	authorDID         string
	labels            []moderation.Label
	authorLabels      []moderation.Label
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

func NewFeedViewFromItems(items []bluesky.FeedItem, cursor string) FeedView {
	views := make([]PostView, 0, len(items))
	for _, item := range items {
		views = append(views, NewPostViewFromFeedItem(item))
	}
	return FeedView{
		Posts:      views,
		NextCursor: cursor,
	}
}

func NewPostViewFromFeedItem(item bluesky.FeedItem) PostView {
	view := NewPostView(item.Post)
	if item.Reason != nil && item.Reason.RepostedBy.Handle != "" {
		repostedBy := authorView(item.Reason.RepostedBy)
		view.RepostedByMaybe = &repostedBy
	}
	if item.Reply != nil && item.Reply.Parent != nil {
		parent := insetPostView(NewPostView(*item.Reply.Parent))
		view.ReplyParentMaybe = &parent
		view.replyParentURI = ""
	}
	return view
}

func authorView(author bluesky.Author) AuthorView {
	return AuthorView{
		Handle:      author.Handle,
		DisplayName: actor.Name(author.DisplayName, author.Handle),
		Avatar:      author.Avatar,
	}
}

func NewPostView(post bluesky.Post) PostView {
	view := PostView{
		ID:                postID(post.URI),
		AuthorHandle:      post.Author.Handle,
		AuthorDisplayName: actor.Name(post.Author.DisplayName, post.Author.Handle),
		AuthorAvatar:      post.Author.Avatar,
		LikeCount:         post.LikeCount,
		RepostCount:       post.RepostCount,
		ReplyCount:        post.ReplyCount,
		Text:              post.Record.Text,
		TextSegments:      richtext.BuildSegments(post.Record.Text, post.Record.Facets),
		CreatedAt:         post.Record.CreatedAt,
		replyParentURI:    post.ReplyParentURI(),
		authorDID:         post.Author.DID,
		labels:            moderationLabels(post.AllLabels()),
		authorLabels:      moderationLabels(post.Author.Labels),
	}
	appendImagesFromEmbed(&view, post.Embed)

	if post.Embed != nil {
		if quoted := post.Embed.QuotedPost(); quoted != nil {
			quotedView := insetPostView(NewPostView(*quoted))
			view.QuotedPostMaybe = &quotedView
		}
	}

	return view
}

func moderationLabels(labels []bluesky.Label) []moderation.Label {
	if len(labels) == 0 {
		return nil
	}
	out := make([]moderation.Label, 0, len(labels))
	for _, label := range labels {
		out = append(out, moderation.Label{Val: label.Val, Src: label.Src})
	}
	return out
}

func postID(uri string) string {
	id, _ := atproto.PostRkey(uri)
	return id
}

func appendImagesFromEmbed(view *PostView, embed *bluesky.Embed) {
	if embed == nil {
		return
	}
	for _, image := range embed.MediaImages() {
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
