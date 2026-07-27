package healthz

import (
	"fmt"

	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/version"
	g "maragu.dev/gomponents"
)

func Preview(publicBaseURL string) g.Node {
	shortID := version.ShortID()
	meta := page.PageMeta{
		Title:        "Twisky is live",
		Description:  fmt.Sprintf("Build %s · Twisky %s", shortID, page.Version),
		CanonicalURL: page.AbsoluteURL(publicBaseURL, "/healthz"),
		OGType:       "website",
	}
	return page.PreviewPage(meta, fmt.Sprintf("ok %s", shortID))
}
