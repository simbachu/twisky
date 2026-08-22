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
	FilterText string
	Blurred    bool
	NoOverride bool
	AlertText  string
	BlurMedia  bool
	BlurAvatar bool
	AvatarText string
}

type ImageView struct {
	Thumb    string
	Fullsize string
	Alt      string
	Width    int
	Height   int
}

type VideoView struct {
	Playlist     string
	Thumbnail    string
	Alt          string
	Width        int
	Height       int
	Presentation string
}

type AuthorView struct {
	Handle      string
	DisplayName string
	Avatar      string
}

type LinkPreviewView struct {
	URI         string
	Title       string
	Description string
	Thumb       string
}

type PostView struct {
	ID                  string
	AuthorHandle        string
	AuthorDisplayName   string
	AuthorAvatar        string
	LikeCount           int
	RepostCount         int
	ReplyCount          int
	Liked               bool
	Reposted            bool
	Text                string
	TextSegments        []richtext.Segment
	CreatedAt           time.Time
	Images              []ImageView
	Videos              []VideoView
	LinkPreviewMaybe    *LinkPreviewView
	RepostedByMaybe     *AuthorView
	ReplyParentMaybe    *PostView
	QuotedPostMaybe     *PostView
	Moderation          ModerationView
	replyParentURI      string
	postURI             string
	postCID             string
	likeURI             string
	repostURI           string
	authorDID           string
	threadRootAuthorDID string
	labels              []moderation.Label
	authorLabels        []moderation.Label
}

type FeedView struct {
	Posts      []PostView
	NextCursor string
}

func (v PostView) AuthorDID() string {
	return v.authorDID
}

func (v PostView) URI() string {
	return v.postURI
}

func (v PostView) CID() string {
	return v.postCID
}

func (v PostView) LikeURI() string {
	return v.likeURI
}

func (v PostView) RepostURI() string {
	return v.repostURI
}

func (v PostView) ThreadRootAuthorDID() string {
	if v.threadRootAuthorDID != "" {
		return v.threadRootAuthorDID
	}
	return v.authorDID
}

// PostViewWithAuthorDID returns view with authorDID set for tests outside this package.
func PostViewWithAuthorDID(view PostView, authorDID string) PostView {
	view.authorDID = authorDID
	return view
}

// PostViewWithStrongRef returns view with URI/CID set for tests outside this package.
func PostViewWithStrongRef(view PostView, uri, cid string) PostView {
	view.postURI = uri
	view.postCID = cid
	return view
}

// PostViewWithEngagement returns view with viewer engagement record URIs set for tests.
func PostViewWithEngagement(view PostView, likeURI, repostURI string) PostView {
	view.likeURI = likeURI
	view.repostURI = repostURI
	return view
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
		parent := InsetPostView(NewPostView(*item.Reply.Parent))
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
		ID:                  postID(post.URI),
		AuthorHandle:        post.Author.Handle,
		AuthorDisplayName:   actor.Name(post.Author.DisplayName, post.Author.Handle),
		AuthorAvatar:        post.Author.Avatar,
		LikeCount:           post.LikeCount,
		RepostCount:         post.RepostCount,
		ReplyCount:          post.ReplyCount,
		Liked:               post.Viewer != nil && post.Viewer.Like != "",
		Reposted:            post.Viewer != nil && post.Viewer.Repost != "",
		Text:                post.Record.Text,
		TextSegments:        richtext.BuildSegments(post.Record.Text, post.Record.Facets),
		CreatedAt:           post.Record.CreatedAt,
		replyParentURI:      post.ReplyParentURI(),
		postURI:             post.URI,
		postCID:             post.CID,
		authorDID:           post.Author.DID,
		threadRootAuthorDID: threadRootAuthorDIDFromPost(post),
		labels:              moderationLabels(post.AllLabels()),
		authorLabels:        moderationLabels(post.Author.Labels),
	}
	if post.Viewer != nil {
		view.likeURI = post.Viewer.Like
		view.repostURI = post.Viewer.Repost
	}
	appendImagesFromEmbed(&view, post.Embed)
	appendVideosFromEmbed(&view, post.Embed)
	appendLinkPreviewFromEmbed(&view, post.Embed)

	if post.Embed != nil {
		if quoted := post.Embed.QuotedPost(); quoted != nil {
			quotedView := InsetPostView(NewPostView(*quoted))
			view.QuotedPostMaybe = &quotedView
		}
	}

	return view
}

func threadRootAuthorDIDFromPost(post bluesky.Post) string {
	if post.Record.Reply != nil && post.Record.Reply.Root.URI != "" {
		if did, err := atproto.PostAuthorDID(post.Record.Reply.Root.URI); err == nil {
			return did
		}
	}
	return post.Author.DID
}

func moderationLabels(labels []bluesky.Label) []moderation.Label {
	if len(labels) == 0 {
		return nil
	}
	out := make([]moderation.Label, 0, len(labels))
	for _, label := range labels {
		out = append(out, moderation.Label{Val: label.Val, Src: label.Src, URI: label.URI})
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

func appendVideosFromEmbed(view *PostView, embed *bluesky.Embed) {
	if embed == nil {
		return
	}
	for _, video := range embed.MediaVideos() {
		videoView := VideoView{
			Playlist:     video.Playlist,
			Thumbnail:    video.Thumbnail,
			Alt:          video.Alt,
			Presentation: video.Presentation,
		}
		if video.AspectRatio != nil {
			videoView.Width = video.AspectRatio.Width
			videoView.Height = video.AspectRatio.Height
		}
		view.Videos = append(view.Videos, videoView)
	}
}

func appendLinkPreviewFromEmbed(view *PostView, embed *bluesky.Embed) {
	if embed == nil {
		return
	}
	external := embed.ExternalLink()
	if external == nil || external.URI == "" {
		return
	}
	view.LinkPreviewMaybe = &LinkPreviewView{
		URI:         external.URI,
		Title:       external.Title,
		Description: external.Description,
		Thumb:       external.Thumb,
	}
}
