package post

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/simbachu/twisky/internal/components/page"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/richtext"
)

type PreviewImage struct {
	URL       string
	LargeCard bool
	Alt       string
	Width     int
	Height    int
}

func postPageMeta(view feedquery.PostPageView, publicBaseURL string) page.PageMeta {
	post := view.Post
	byline := postAuthorByline(post)
	title := byline
	description := ""
	image := PreviewImage{}

	if post.Moderation.Filtered {
		description = "Post hidden by moderation on Twisky"
	} else {
		if view.ReplyParentMaybe != nil {
			title = "Reply by " + byline
		}
		description = postDescription(view)
		image = PreviewImageFromPost(post)
		if image.Alt == "" {
			image.Alt = byline + " on Twisky"
		}
	}

	return page.PageMeta{
		Title:          title,
		Description:    description,
		CanonicalURL:   page.AbsoluteURL(publicBaseURL, "/"+post.AuthorHandle+"/post/"+url.PathEscape(post.ID)),
		ImageURL:       image.URL,
		OGType:         "article",
		LargeImageCard: image.LargeCard,
		PublishedTime:  post.CreatedAt,
		AuthorURL:      page.AbsoluteURL(publicBaseURL, "/"+post.AuthorHandle),
		AuthorHandle:   post.AuthorHandle,
		Tags:           postHashtags(post),
		ImageAlt:       image.Alt,
		ImageWidth:     image.Width,
		ImageHeight:    image.Height,
	}
}

func postAuthorByline(post feedquery.PostView) string {
	return fmt.Sprintf("%s (@%s)", post.AuthorDisplayName, post.AuthorHandle)
}

func postDescription(view feedquery.PostPageView) string {
	post := view.Post
	parts := make([]string, 0, 4)

	switch {
	case view.ReplyParentMaybe != nil:
		parts = append(parts, "Replying to @"+view.ReplyParentMaybe.Handle)
	case view.HasAncestors:
		parts = append(parts, "Reply in thread")
	}

	text := strings.TrimSpace(post.Text)
	if text != "" {
		parts = append(parts, text)
	} else if post.LinkPreviewMaybe == nil && post.QuotedPostMaybe == nil {
		parts = append(parts, fmt.Sprintf("Post by %s on Twisky", post.AuthorDisplayName))
	}

	if post.QuotedPostMaybe != nil {
		quoted := post.QuotedPostMaybe
		excerpt := page.TruncateDescription(quoted.Text, 80)
		if excerpt == "" {
			parts = append(parts, "Quoting @"+quoted.AuthorHandle)
		} else {
			parts = append(parts, "Quoting @"+quoted.AuthorHandle+": "+excerpt)
		}
	}

	if post.LinkPreviewMaybe != nil {
		linkLabel := strings.TrimSpace(post.LinkPreviewMaybe.Title)
		if linkLabel == "" {
			linkLabel = strings.TrimSpace(post.LinkPreviewMaybe.URI)
		}
		if linkLabel != "" && (text == "" || len([]rune(text)) < 40) {
			parts = append(parts, linkLabel)
		}
	}

	if post.ReplyCount > 0 {
		label := strconv.Itoa(post.ReplyCount) + " replies"
		if post.ReplyCount == 1 {
			label = "1 reply"
		}
		parts = append(parts, label)
	}

	if len(parts) == 0 {
		return fmt.Sprintf("Post by %s on Twisky", post.AuthorDisplayName)
	}
	return page.TruncateDescription(strings.Join(parts, " · "), 200)
}

func postHashtags(post feedquery.PostView) []string {
	seen := make(map[string]struct{})
	tags := make([]string, 0)
	for _, segment := range post.TextSegments {
		if segment.Kind != richtext.Tag || segment.Tag == "" {
			continue
		}
		if _, ok := seen[segment.Tag]; ok {
			continue
		}
		seen[segment.Tag] = struct{}{}
		tags = append(tags, segment.Tag)
	}
	return tags
}

func PreviewImageFromPost(post feedquery.PostView) PreviewImage {
	if len(post.Images) > 0 {
		img := post.Images[0]
		url := img.Fullsize
		if url == "" {
			url = img.Thumb
		}
		if url != "" {
			return PreviewImage{
				URL:       url,
				LargeCard: true,
				Alt:       img.Alt,
				Width:     img.Width,
				Height:    img.Height,
			}
		}
	}
	if post.LinkPreviewMaybe != nil && post.LinkPreviewMaybe.Thumb != "" {
		return PreviewImage{
			URL:       post.LinkPreviewMaybe.Thumb,
			LargeCard: true,
			Alt:       strings.TrimSpace(post.LinkPreviewMaybe.Title),
		}
	}
	if len(post.Videos) > 0 && post.Videos[0].Thumbnail != "" {
		video := post.Videos[0]
		return PreviewImage{
			URL:       video.Thumbnail,
			LargeCard: true,
			Alt:       video.Alt,
			Width:     video.Width,
			Height:    video.Height,
		}
	}
	if post.QuotedPostMaybe != nil {
		if quoted := PreviewImageFromPost(*post.QuotedPostMaybe); quoted.URL != "" {
			return quoted
		}
	}
	if post.AuthorAvatar != "" {
		return PreviewImage{
			URL:       post.AuthorAvatar,
			LargeCard: true,
		}
	}
	return PreviewImage{}
}
