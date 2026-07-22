package tag

import (
	"fmt"
	"net/url"

	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	tagquery "github.com/simbachu/twisky/internal/query/tag"
)

func tagPageMeta(view tagquery.TagView, publicBaseURL string) page.PageMeta {
	tagName := "#" + view.Tag
	image := tagPreviewImage(view.Feed.Posts)
	return page.PageMeta{
		Title:       tagName + " on Twisky",
		Description: fmt.Sprintf("Recent posts tagged with %s on Twisky", tagName),
		CanonicalURL: page.AbsoluteURL(
			publicBaseURL,
			"/tagged/"+url.PathEscape(view.Tag),
		),
		ImageURL:       image.URL,
		OGType:         "website",
		LargeImageCard: image.LargeCard,
		ImageAlt:       image.Alt,
		ImageWidth:     image.Width,
		ImageHeight:    image.Height,
	}
}

func tagPreviewImage(posts []feedquery.PostView) post.PreviewImage {
	for _, candidate := range posts {
		if candidate.Moderation.Filtered {
			continue
		}
		image := post.PreviewImageFromPost(candidate)
		if image.URL != "" {
			if image.Alt == "" {
				image.Alt = fmt.Sprintf("%s (@%s) on Twisky", candidate.AuthorDisplayName, candidate.AuthorHandle)
			}
			return image
		}
	}
	return post.PreviewImage{}
}
