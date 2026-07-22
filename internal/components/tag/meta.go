package tag

import (
	"fmt"
	"net/url"

	"github.com/simbachu/twisky/internal/components/page"
	tagquery "github.com/simbachu/twisky/internal/query/tag"
)

func tagPageMeta(view tagquery.TagView, publicBaseURL string) page.PageMeta {
	tagName := "#" + view.Tag
	return page.PageMeta{
		Title:       tagName + " on Twisky",
		Description: fmt.Sprintf("Recent posts tagged with %s on Twisky", tagName),
		CanonicalURL: page.AbsoluteURL(
			publicBaseURL,
			"/tagged/"+url.PathEscape(view.Tag),
		),
		OGType: "website",
	}
}
