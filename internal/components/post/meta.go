package post

import (
	"fmt"
	"net/url"

	"github.com/simbachu/twisky/internal/components/page"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func postPageMeta(view feedquery.PostPageView, publicBaseURL string) page.PageMeta {
	post := view.Post
	title := page.TruncateDescription(post.Text, 80)
	description := page.TruncateDescription(post.Text, 200)
	largeImageCard := false
	imageURL := ""

	if post.Moderation.Filtered {
		title = "Post by " + post.AuthorDisplayName
		description = "Post hidden by moderation on Twisky"
	} else {
		if title == "" {
			title = "Post by " + post.AuthorDisplayName
		}
		if description == "" {
			description = fmt.Sprintf("Post by %s on Twisky", post.AuthorDisplayName)
		}
		imageURL, largeImageCard = postPreviewImage(post)
	}

	return page.PageMeta{
		Title:          title,
		Description:    description,
		CanonicalURL:   page.AbsoluteURL(publicBaseURL, "/"+post.AuthorHandle+"/post/"+url.PathEscape(post.ID)),
		ImageURL:       imageURL,
		OGType:         "article",
		LargeImageCard: largeImageCard,
	}
}

func postPreviewImage(post feedquery.PostView) (imageURL string, largeImage bool) {
	if len(post.Images) > 0 {
		if post.Images[0].Fullsize != "" {
			return post.Images[0].Fullsize, true
		}
		if post.Images[0].Thumb != "" {
			return post.Images[0].Thumb, true
		}
	}
	if post.LinkPreviewMaybe != nil && post.LinkPreviewMaybe.Thumb != "" {
		return post.LinkPreviewMaybe.Thumb, true
	}
	if len(post.Videos) > 0 && post.Videos[0].Thumbnail != "" {
		return post.Videos[0].Thumbnail, true
	}
	if post.AuthorAvatar != "" {
		return post.AuthorAvatar, false
	}
	return "", false
}
