package post

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/simbachu/twisky/internal/components/ui"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Post(view feedquery.PostView, now time.Time) g.Node {
	return PostArticle(view, now, "post")
}

func PostArticle(view feedquery.PostView, now time.Time, class string, extra ...g.Node) g.Node {
	if view.Moderation.Filtered {
		return nil
	}

	repostedBy, replyParent, replyParentID := postHeaderMeta(view)
	children := []g.Node{
		ui.PostHeader(authorInfo(view), view.CreatedAt, now, repostedBy, replyParent, replyParentID),
		moderationNotice(view.Moderation),
		moderationBody(view, now),
		postFooter(view),
	}
	children = append(children, extra...)
	return Article(g.Attr("class", class), g.Attr("id", "post-"+url.PathEscape(view.ID)), g.Group(children))
}

func postFooter(view feedquery.PostView) g.Node {
	return Footer(
		Nav(
			g.Attr("aria-label", "Post actions"),
			ui.SegmentedGroup("Engagement actions",
				ui.PostEngagement(ui.IconReply, "Reply", view.ReplyCount),
				ui.PostEngagement(ui.IconRepost, "Repost", view.RepostCount),
				ui.PostEngagement(ui.IconLike, "Like", view.LikeCount),
			),
			ui.SegmentedGroup("Bookmark", ui.PostEngagement(ui.IconBookmark, "Bookmark", 0)),
			ui.SegmentedGroup("Share", ui.PostEngagement(ui.IconShare, "Share", 0)),
			ui.SegmentedGroup("More options", ui.PostEngagement(ui.IconMore, "More options", 0)),
		),
	)
}

func quotedInset(maybe *feedquery.PostView, now time.Time) g.Node {
	if maybe == nil {
		return nil
	}
	return ClickableInset(maybe, now, "View quoted post")
}

// ClickableInset wraps an inset post with an overlay link to the post page.
func ClickableInset(view *feedquery.PostView, now time.Time, ariaLabel string) g.Node {
	if view == nil || view.Moderation.Filtered {
		return nil
	}
	href := "/" + view.AuthorHandle + "/post/" + url.PathEscape(view.ID)
	return Div(
		g.Attr("class", "clickable-inset"),
		A(
			g.Attr("href", href),
			g.Attr("aria-label", ariaLabel),
		),
		Div(InsetPost(view, now)),
	)
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
		postLinkPreview(view.LinkPreviewMaybe, view.Moderation),
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
	return P(ui.RichText(view.TextSegments))
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

func postLinkPreview(preview *feedquery.LinkPreviewView, mod feedquery.ModerationView) g.Node {
	card := postLinkPreviewBase(preview)
	if card == nil {
		return nil
	}
	if preview.Thumb == "" || mod.Blurred || !mod.BlurMedia {
		return card
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
		card,
	)
}

func postLinkPreviewBase(preview *feedquery.LinkPreviewView) g.Node {
	if preview == nil || preview.URI == "" {
		return nil
	}

	body := []g.Node{
		Span(g.Attr("class", "post-link-preview-title"), g.Text(preview.Title)),
	}
	if preview.Description != "" {
		body = append(body, Span(g.Attr("class", "post-link-preview-description"), g.Text(preview.Description)))
	}
	if host := linkPreviewHost(preview.URI); host != "" {
		body = append(body, Span(g.Attr("class", "post-link-preview-host"), g.Text(host)))
	}

	children := []g.Node{Span(g.Attr("class", "post-link-preview-body"), g.Group(body))}
	if preview.Thumb != "" {
		children = append([]g.Node{
			Img(
				g.Attr("class", "post-link-preview-thumb"),
				g.Attr("src", preview.Thumb),
				g.Attr("alt", ""),
				g.Attr("loading", "lazy"),
			),
		}, children...)
	}

	return A(
		g.Attr("class", "post-link-preview"),
		g.Attr("href", preview.URI),
		g.Attr("target", "_blank"),
		g.Attr("rel", "noopener noreferrer"),
		g.Attr("style", "pointer-events: auto"),
		g.Group(children),
	)
}

func linkPreviewHost(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return ""
	}
	return parsed.Host
}
