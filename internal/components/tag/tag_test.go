package tag_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/tag"
	tagquery "github.com/simbachu/twisky/internal/query/tag"
)

func TestTag_RendersSocialMetaTags(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := tag.Tag(tagquery.TagView{
		Tag: "art",
	}, time.Now().UTC(), nil, "https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:title" content="#art on Twisky"`,
		`property="og:description" content="Recent posts tagged with #art on Twisky"`,
		`property="og:type" content="website"`,
		`rel="canonical" href="https://twisky.test/tagged/art"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, `property="og:image"`) {
		t.Fatalf("html = %q, want no og:image for tag page", html)
	}
}
