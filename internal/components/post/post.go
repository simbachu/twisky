package post

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/richtext"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Post(view feedquery.PostView, now time.Time) g.Node {
	if view.Moderation.Filtered {
		return nil
	}

	repostedBy, replyParent, replyParentID := postHeaderMeta(view)
	return Article(g.Attr("class", "post"), g.Attr("id", "post-"+url.PathEscape(view.ID)),
		ui.PostHeader(authorInfo(view), view.CreatedAt, now, repostedBy, replyParent, replyParentID),
		moderationNotice(view.Moderation),
		moderationBody(view, now),
		Footer(
			Nav(
				Ul(
					Li(ui.ActionButton(ui.IconReply, "Reply", view.ReplyCount)),
					Li(ui.ActionButton(ui.IconRepost, "Repost", view.RepostCount)),
					Li(ui.ActionButton(ui.IconLike, "Like", view.LikeCount)),
				),
			),
		),
	)
}

func quotedInset(maybe *feedquery.PostView, now time.Time) g.Node {
	if maybe == nil {
		return nil
	}
	return InsetPost(maybe, now)
}

// Condensed post view with Author, Icon, Text and images
func InsetPost(view *feedquery.PostView, now time.Time) g.Node {
	if view == nil || view.Moderation.Filtered {
		return nil
	}
	return Article(g.Attr("class", "post inset-post"), g.Attr("id", "inset-post-"+url.PathEscape(view.ID)),
		ui.PostHeader(authorInfo(*view), view.CreatedAt, now, nil, nil, ""),
		moderationNotice(view.Moderation),
		moderationBody(*view, now),
	)
}

func moderationNotice(mod feedquery.ModerationView) g.Node {
	if mod.AlertText == "" || mod.Blurred {
		return nil
	}
	return Div(g.Attr("class", "post-moderation-notice"), g.Text(mod.AlertText))
}

func moderationBody(view feedquery.PostView, now time.Time) g.Node {
	content := postContent(view, now)
	if !view.Moderation.Blurred {
		return content
	}
	if view.Moderation.NoOverride {
		return Div(g.Attr("class", "post-moderation-gate"),
			moderationCover(view.Moderation),
		)
	}
	return Details(g.Attr("class", "post-moderation-gate"),
		Summary(
			moderationCover(view.Moderation),
			g.Text("Show anyway"),
		),
		Div(content),
	)
}

func moderationCover(mod feedquery.ModerationView) g.Node {
	message := mod.AlertText
	if message == "" {
		message = "Content warning"
	}
	return P(g.Text(message))
}

func postContent(view feedquery.PostView, now time.Time) g.Node {
	return g.Group{
		postText(view),
		postFigure(view.Images, view.Moderation),
		postVideo(view.Videos, view.Moderation),
		quotedInset(view.QuotedPostMaybe, now),
	}
}

func authorInfo(view feedquery.PostView) ui.AuthorInfo {
	return ui.AuthorInfo{
		Handle:      view.AuthorHandle,
		DisplayName: view.AuthorDisplayName,
		Avatar:      view.AuthorAvatar,
	}
}

func postHeaderMeta(view feedquery.PostView) (repostedBy, replyParent *ui.AuthorInfo, replyParentID string) {
	if view.RepostedByMaybe != nil {
		info := authorInfoFromView(*view.RepostedByMaybe)
		repostedBy = &info
	}
	if view.ReplyParentMaybe != nil {
		info := authorInfo(*view.ReplyParentMaybe)
		replyParent = &info
		replyParentID = view.ReplyParentMaybe.ID
	}
	return repostedBy, replyParent, replyParentID
}

func authorInfoFromView(view feedquery.AuthorView) ui.AuthorInfo {
	return ui.AuthorInfo{
		Handle:      view.Handle,
		DisplayName: view.DisplayName,
		Avatar:      view.Avatar,
	}
}

func postText(view feedquery.PostView) g.Node {
	if len(view.TextSegments) == 0 {
		return P(g.Text(view.Text))
	}
	return P(g.Group(g.Map(view.TextSegments, segmentNode)))
}

func segmentNode(segment richtext.Segment) g.Node {
	switch segment.Kind {
	case richtext.Tag:
		return A(
			g.Attr("href", "/tagged/"+url.PathEscape(segment.Tag)),
			g.Attr("style", "pointer-events: auto"),
			Span(g.Attr("class", "facet-tag"), g.Text(segment.Text)),
		)
	case richtext.Mention:
		return A(
			g.Attr("href", "/"+url.PathEscape(segment.Mention)),
			g.Attr("style", "pointer-events: auto"),
			Span(g.Attr("class", "facet-mention"), g.Text(segment.Text)),
		)
	case richtext.Link:
		return A(
			g.Attr("href", segment.URI),
			g.Attr("target", "_blank"),
			g.Attr("rel", "noopener noreferrer"),
			g.Attr("style", "pointer-events: auto"),
			Span(g.Attr("class", "facet-link"), g.Text(segment.Text)),
		)
	default:
		return g.Text(segment.Text)
	}
}

func postFigure(images []feedquery.ImageView, mod feedquery.ModerationView) g.Node {
	figure := postFigureBase(images)
	if figure == nil {
		return nil
	}
	if mod.Blurred || !mod.BlurMedia {
		return figure
	}
	if mod.NoOverride {
		return Div(g.Attr("class", "post-moderation-gate"),
			moderationCover(mod),
		)
	}
	return Details(g.Attr("class", "post-moderation-gate"),
		Summary(
			moderationCover(mod),
			g.Text("Show media"),
		),
		figure,
	)
}

func postFigureBase(images []feedquery.ImageView) g.Node {
	if len(images) == 0 {
		return nil
	}
	count := len(images)
	if count > 4 {
		count = 4
	}
	return Figure(
		g.Attr("class", fmt.Sprintf("post-images post-images-%d", count)),
		g.Group(g.Map(images, postImage)),
	)
}

func postImage(image feedquery.ImageView) g.Node {
	attrs := []g.Node{
		g.Attr("src", image.Thumb),
		g.Attr("alt", image.Alt),
		g.Attr("loading", "lazy"),
	}
	if image.Fullsize != "" {
		attrs = append(attrs, g.Attr("srcset", image.Thumb+" 1000w, "+image.Fullsize+" 2000w"))
	}
	if image.Width > 0 && image.Height > 0 {
		attrs = append(attrs,
			g.Attr("width", strconv.Itoa(image.Width)),
			g.Attr("height", strconv.Itoa(image.Height)),
		)
	}
	return Img(attrs...)
}

func postVideo(videos []feedquery.VideoView, mod feedquery.ModerationView) g.Node {
	figure := postVideoBase(videos)
	if figure == nil {
		return nil
	}
	if mod.Blurred || !mod.BlurMedia {
		return figure
	}
	if mod.NoOverride {
		return Div(g.Attr("class", "post-moderation-gate"),
			moderationCover(mod),
		)
	}
	return Details(g.Attr("class", "post-moderation-gate"),
		Summary(
			moderationCover(mod),
			g.Text("Show media"),
		),
		figure,
	)
}

func postVideoBase(videos []feedquery.VideoView) g.Node {
	if len(videos) == 0 {
		return nil
	}
	video := videos[0]
	attrs := []g.Node{
		g.Attr("class", "post-video-player"),
		g.Attr("poster", video.Thumbnail),
		g.Attr("preload", "none"),
		g.Attr("playsinline", ""),
		g.Attr("data-playlist", video.Playlist),
		g.Attr("data-presentation", video.Presentation),
	}
	if video.Alt != "" {
		attrs = append(attrs, g.Attr("aria-label", video.Alt))
	}
	if video.Width > 0 && video.Height > 0 {
		attrs = append(attrs,
			g.Attr("width", strconv.Itoa(video.Width)),
			g.Attr("height", strconv.Itoa(video.Height)),
		)
	}
	if video.Presentation == "gif" {
		attrs = append(attrs,
			g.Attr("autoplay", ""),
			g.Attr("loop", ""),
			g.Attr("muted", ""),
			g.Attr("src", video.Playlist),
		)
	}
	nodes := []g.Node{Video(attrs...)}
	if video.Presentation != "gif" {
		nodes = append(nodes, Span(
			g.Attr("class", "post-video-play"),
			g.Attr("aria-hidden", "true"),
			g.Text("▶️"),
		))
	}
	return Figure(
		g.Attr("class", "post-video"),
		g.Group(nodes),
	)
}
