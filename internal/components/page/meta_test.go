package page_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/page"
)

func TestTruncateDescription_TruncatesLongText(t *testing.T) {
	t.Parallel()

	got := page.TruncateDescription(strings.Repeat("a", 250), 200)
	if len([]rune(got)) != 201 {
		t.Fatalf("len = %d, want 201 including ellipsis", len([]rune(got)))
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("TruncateDescription() = %q, want ellipsis suffix", got)
	}
}

func TestTruncateDescription_LeavesShortTextUntouched(t *testing.T) {
	t.Parallel()

	got := page.TruncateDescription("hello world", 200)
	if got != "hello world" {
		t.Fatalf("TruncateDescription() = %q, want hello world", got)
	}
}

func TestAbsoluteURL_JoinsBaseAndPath(t *testing.T) {
	t.Parallel()

	got := page.AbsoluteURL("https://dev.twisky.app/", "/simbachu.com")
	if got != "https://dev.twisky.app/simbachu.com" {
		t.Fatalf("AbsoluteURL() = %q", got)
	}
}

func TestAbsoluteURL_ReturnsEmptyWhenBaseMissing(t *testing.T) {
	t.Parallel()

	if got := page.AbsoluteURL("", "/simbachu.com"); got != "" {
		t.Fatalf("AbsoluteURL() = %q, want empty", got)
	}
}

func TestPage_RendersSocialMetaTags(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	meta := page.PageMeta{
		Title:          "Spectral (@simbachu.com)",
		Description:    "gamer, mother, husband",
		CanonicalURL:   "https://twisky.test/simbachu.com",
		ImageURL:       "https://cdn.example/avatar.jpg",
		OGType:         "profile",
		LargeImageCard: false,
	}
	if err := page.Page(meta, nil).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:title" content="Spectral (@simbachu.com)"`,
		`property="og:description" content="gamer, mother, husband"`,
		`property="og:site_name" content="Twisky"`,
		`property="og:type" content="profile"`,
		`property="og:url" content="https://twisky.test/simbachu.com"`,
		`rel="canonical" href="https://twisky.test/simbachu.com"`,
		`property="og:image" content="https://cdn.example/avatar.jpg"`,
		`name="twitter:card" content="summary"`,
		`name="twitter:image" content="https://cdn.example/avatar.jpg"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestPage_OmitsCanonicalAndImageWhenUnset(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	meta := page.PageMeta{
		Title:       "Tag",
		Description: "Recent posts",
		OGType:      "website",
	}
	if err := page.Page(meta, nil).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `rel="canonical"`) {
		t.Fatalf("html = %q, want no canonical link", html)
	}
	if strings.Contains(html, `property="og:url"`) {
		t.Fatalf("html = %q, want no og:url", html)
	}
	if strings.Contains(html, `property="og:image"`) {
		t.Fatalf("html = %q, want no og:image", html)
	}
	if !strings.Contains(html, `name="twitter:card" content="summary"`) {
		t.Fatalf("html = %q, want summary twitter card", html)
	}
}

func TestPage_UsesLargeImageTwitterCard(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	meta := page.PageMeta{
		Title:          "Post",
		Description:    "hello",
		ImageURL:       "https://cdn.example/post.jpg",
		OGType:         "article",
		LargeImageCard: true,
	}
	if err := page.Page(meta, nil).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `name="twitter:card" content="summary_large_image"`) {
		t.Fatalf("html = %q, want summary_large_image twitter card", html)
	}
}
