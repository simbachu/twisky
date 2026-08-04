package ui_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/components/ui"
)

func TestActionableAvatar_PeekableRendersShieldFallbackForLabeler(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionableAvatar(ui.AuthorInfo{
		Handle:      "labeler.example.com",
		DisplayName: "Example Labeler Service",
		IsLabeler:   true,
	}, ui.AuthorPeekable).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="profile-icon-emoji"`) {
		t.Fatalf("html = %q, want emoji avatar", html)
	}
	if !strings.Contains(html, actor.AvatarEmojiShield) {
		t.Fatalf("html = %q, want shield emoji", html)
	}
	if !strings.Contains(html, `data-kind="labeler"`) {
		t.Fatalf("html = %q, want data-kind=labeler", html)
	}
	if !strings.Contains(html, `href="/labeler.example.com"`) {
		t.Fatalf("html = %q, want profile href", html)
	}
	if !strings.Contains(html, `data-author-interactive="peekable"`) {
		t.Fatalf("html = %q, want peekable marker", html)
	}
}

func TestActionableAvatar_StaticOmitsLinkAndPeekMarker(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionableAvatar(ui.AuthorInfo{
		Handle:      "dev.example",
		DisplayName: "Dev",
		Avatar:      "https://cdn.example/a.jpg",
	}, ui.AuthorStatic).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `href=`) {
		t.Fatalf("html = %q, want no href for static", html)
	}
	if strings.Contains(html, `data-author-interactive=`) {
		t.Fatalf("html = %q, want no interactive marker for static", html)
	}
	if !strings.Contains(html, `class="profile-icon"`) {
		t.Fatalf("html = %q, want profile-icon", html)
	}
}

func TestActionableAvatar_LinkedMarksWithoutPeek(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionableAvatar(ui.AuthorInfo{
		Handle:      "dev.example",
		DisplayName: "Dev",
		Avatar:      "https://cdn.example/a.jpg",
	}, ui.AuthorLinked).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `href="/dev.example"`) {
		t.Fatalf("html = %q, want profile href", html)
	}
	if !strings.Contains(html, `data-author-interactive="linked"`) {
		t.Fatalf("html = %q, want linked marker", html)
	}
	if strings.Contains(html, `data-author-interactive="peekable"`) {
		t.Fatalf("html = %q, want no peekable marker", html)
	}
}

func TestActionableAvatar_OmitsDataKindForNonLabeler(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionableAvatar(ui.AuthorInfo{
		Handle:      "dev.example",
		DisplayName: "Dev",
		Avatar:      "https://cdn.example/a.jpg",
	}, ui.AuthorPeekable).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	if strings.Contains(buf.String(), `data-kind=`) {
		t.Fatalf("html = %q, want no data-kind", buf.String())
	}
}

func TestActionableAvatar_LabelerKindOnModeratedPlaceholder(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionableAvatar(ui.AuthorInfo{
		Handle:     "labeler.example.com",
		IsLabeler:  true,
		BlurAvatar: true,
	}, ui.AuthorPeekable).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `data-kind="labeler"`) {
		t.Fatalf("html = %q, want data-kind=labeler on moderated avatar", html)
	}
	if !strings.Contains(html, `profile-icon-moderated`) {
		t.Fatalf("html = %q, want moderated avatar", html)
	}
}

func TestActionableAvatar_RendersButterflyFallbackForBskyApp(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ActionableAvatar(ui.AuthorInfo{
		Handle:      "user.bsky.app",
		DisplayName: "Bluesky User",
	}, ui.AuthorPeekable).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, actor.AvatarEmojiButterfly) {
		t.Fatalf("html = %q, want butterfly emoji", html)
	}
}

func TestAuthorName_PeekableIsLink(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.AuthorName(ui.AuthorInfo{
		Handle:      "dev.example",
		DisplayName: "Dev User",
	}, ui.AuthorPeekable).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.HasPrefix(html, "<a ") {
		t.Fatalf("html = %q, want link", html)
	}
	if !strings.Contains(html, `data-author-interactive="peekable"`) {
		t.Fatalf("html = %q, want peekable marker", html)
	}
	if !strings.Contains(html, "Dev User") {
		t.Fatalf("html = %q, want display name", html)
	}
}

func TestAuthorHandle_StaticIsSpan(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.AuthorHandle(ui.AuthorInfo{
		Handle: "dev.example",
	}, ui.AuthorStatic).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.HasPrefix(html, "<span ") {
		t.Fatalf("html = %q, want span", html)
	}
	if strings.Contains(html, `href=`) {
		t.Fatalf("html = %q, want no href", html)
	}
	if !strings.Contains(html, "@dev.example") {
		t.Fatalf("html = %q, want handle", html)
	}
}

func TestAuthorLink_ShowsNameAndHandle(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.AuthorLink(ui.AuthorInfo{
		Handle:      "dev.example",
		DisplayName: "Dev User",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="author-name"`) {
		t.Fatalf("html = %q, want author-name", html)
	}
	if !strings.Contains(html, "Dev User") {
		t.Fatalf("html = %q, want display name", html)
	}
	if !strings.Contains(html, "@dev.example") {
		t.Fatalf("html = %q, want handle", html)
	}
	if strings.Count(html, `data-author-interactive="peekable"`) != 2 {
		t.Fatalf("html = %q, want peekable name and handle", html)
	}
}

func TestAuthorLink_OmitsNameWhenSameAsHandle(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.AuthorLink(ui.AuthorInfo{
		Handle:      "dev.example",
		DisplayName: "dev.example",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `class="author-name"`) {
		t.Fatalf("html = %q, want no author-name when display equals handle", html)
	}
	if !strings.Contains(html, "@dev.example") {
		t.Fatalf("html = %q, want handle only", html)
	}
}

func TestTimestamp_RendersRelativeWithAbsoluteTitle(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 12, 18, 43, 28, 0, time.UTC)
	now := time.Date(2026, 3, 12, 20, 43, 28, 0, time.UTC)

	var buf bytes.Buffer
	if err := ui.Timestamp(createdAt, now).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `datetime="2026-03-12T18:43:28Z"`) {
		t.Fatalf("html = %q, want datetime attribute", html)
	}
	if !strings.Contains(html, `title="Mar 12, 2026, 6:43 PM UTC"`) {
		t.Fatalf("html = %q, want absolute title", html)
	}
	if !strings.Contains(html, ">2h<") {
		t.Fatalf("html = %q, want relative visible text", html)
	}
}

func TestTimestamp_OmitsWhenZero(t *testing.T) {
	t.Parallel()

	if ui.Timestamp(time.Time{}, time.Now()) != nil {
		t.Fatal("Timestamp() = node, want nil for zero CreatedAt")
	}
}
