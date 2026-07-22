package page_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

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

func TestPage_RendersArticleAndImageStructuredMeta(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	meta := page.PageMeta{
		Title:          "Bluesky (@bsky.app)",
		Description:    "hello from the feed",
		CanonicalURL:   "https://twisky.test/bsky.app/post/abc",
		ImageURL:       "https://cdn.example/post.jpg",
		OGType:         "article",
		PublishedTime:  time.Date(2026, 7, 22, 10, 30, 0, 0, time.UTC),
		AuthorURL:      "https://twisky.test/bsky.app",
		AuthorHandle:   "bsky.app",
		Tags:           []string{"bluesky", "art"},
		ImageAlt:       "Bluesky (@bsky.app) on Twisky",
		ImageWidth:     1200,
		ImageHeight:    675,
		LargeImageCard: true,
	}
	if err := page.Page(meta, nil).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="og:locale" content="en_US"`,
		`property="article:published_time" content="2026-07-22T10:30:00Z"`,
		`property="article:author" content="https://twisky.test/bsky.app"`,
		`property="article:tag" content="bluesky"`,
		`property="article:tag" content="art"`,
		`name="twitter:creator" content="@bsky.app"`,
		`property="og:image:width" content="1200"`,
		`property="og:image:height" content="675"`,
		`property="og:image:alt" content="Bluesky (@bsky.app) on Twisky"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
	if strings.Contains(html, `property="profile:username"`) {
		t.Fatalf("html = %q, want no profile:username on article pages", html)
	}
}

func TestPage_RendersProfileUsername(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	meta := page.PageMeta{
		Title:        "Spectral (@simbachu.com)",
		Description:  "gamer, mother, husband",
		ImageURL:     "https://cdn.example/avatar.jpg",
		OGType:       "profile",
		AuthorHandle: "simbachu.com",
		ImageAlt:     "Spectral (@simbachu.com)",
	}
	if err := page.Page(meta, nil).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`property="profile:username" content="simbachu.com"`,
		`name="twitter:creator" content="@simbachu.com"`,
		`property="og:image:alt" content="Spectral (@simbachu.com)"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %s", html, want)
		}
	}
}

func TestPage_OmitsOptionalArticleFieldsWhenUnset(t *testing.T) {
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
	for _, unwanted := range []string{
		`property="article:published_time"`,
		`property="article:author"`,
		`property="article:tag"`,
		`name="twitter:creator"`,
		`property="profile:username"`,
		`property="og:image:width"`,
		`property="og:image:height"`,
		`property="og:image:alt"`,
	} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("html = %q, want no %s", html, unwanted)
		}
	}
	if !strings.Contains(html, `property="og:locale" content="en_US"`) {
		t.Fatalf("html = %q, want og:locale", html)
	}
}
