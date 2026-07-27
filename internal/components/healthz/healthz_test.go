package healthz_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/healthz"
	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/version"
)

func TestPreview_RendersBuildAndVersionMeta(t *testing.T) {
	t.Parallel()

	prev := version.BuildID
	t.Cleanup(func() { version.BuildID = prev })
	version.BuildID = "9c8a405abcdef"

	var buf bytes.Buffer
	if err := healthz.Preview("https://twisky.test").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:title" content="Twisky is live"`,
		`property="og:description" content="Build 9c8a405 · Twisky ` + page.Version + `"`,
		`property="og:url" content="https://twisky.test/healthz"`,
		`ok 9c8a405`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}
