package ui_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestRepostMeta_RendersHandle(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.RepostMeta(ui.AuthorInfo{
		Handle:      "reposter.example",
		DisplayName: "Reposter",
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="post-meta"`) {
		t.Fatalf("html = %q, want repost meta class", html)
	}
	if !strings.Contains(html, "Reposted by") {
		t.Fatalf("html = %q, want Reposted by", html)
	}
	if !strings.Contains(html, "@reposter.example") {
		t.Fatalf("html = %q, want @reposter.example", html)
	}
}

func TestReplyMeta_RendersHandleAndLink(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.ReplyMeta(ui.AuthorInfo{
		Handle:      "parent.example",
		DisplayName: "Parent",
	}, "parent-id").Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="post-meta post-meta-reply"`) {
		t.Fatalf("html = %q, want reply meta class", html)
	}
	if !strings.Contains(html, "⤷ Reply to @parent.example") {
		t.Fatalf("html = %q, want reply meta text", html)
	}
	if !strings.Contains(html, `href="/parent.example/post/parent-id"`) {
		t.Fatalf("html = %q, want parent post link", html)
	}
}

func TestPostHeader_OmitsMetaWhenNil(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.PostHeader(
		ui.AuthorInfo{Handle: "dev.example", DisplayName: "dev.example"},
		time.Now().UTC(),
		time.Now().UTC(),
		nil,
		nil,
		"",
		true,
	).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, "post-meta-repost") {
		t.Fatalf("html = %q, want no repost meta", html)
	}
	if strings.Contains(html, "post-meta-reply") {
		t.Fatalf("html = %q, want no reply meta", html)
	}
	if !strings.Contains(html, "<header>") {
		t.Fatalf("html = %q, want post header element", html)
	}
}

func TestPostHeader_RendersRepostAndReplyMeta(t *testing.T) {
	t.Parallel()

	repostedBy := ui.AuthorInfo{Handle: "reposter.example", DisplayName: "Reposter"}
	replyParent := ui.AuthorInfo{Handle: "parent.example", DisplayName: "Parent"}

	var buf bytes.Buffer
	if err := ui.PostHeader(
		ui.AuthorInfo{Handle: "dev.example", DisplayName: "Dev"},
		time.Now().UTC(),
		time.Now().UTC(),
		&repostedBy,
		&replyParent,
		"parent-id",
		true,
	).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "Reposted by") {
		t.Fatalf("html = %q, want repost meta", html)
	}
	if !strings.Contains(html, "⤷ Reply to @parent.example") {
		t.Fatalf("html = %q, want reply meta", html)
	}
}

func TestPostHeader_OmitsTimestampWhenDisabled(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 25, 14, 56, 0, 0, time.UTC)
	createdAt := now.Add(-2 * time.Hour)

	var buf bytes.Buffer
	if err := ui.PostHeader(
		ui.AuthorInfo{Handle: "dev.example", DisplayName: "Dev"},
		createdAt,
		now,
		nil,
		nil,
		"",
		false,
	).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if strings.Contains(html, "<time") {
		t.Fatalf("html = %q, want no timestamp in header", html)
	}
}

func TestPostHeader_RendersTimestampWhenEnabled(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 25, 14, 56, 0, 0, time.UTC)
	createdAt := now.Add(-2 * time.Hour)

	var buf bytes.Buffer
	if err := ui.PostHeader(
		ui.AuthorInfo{Handle: "dev.example", DisplayName: "Dev"},
		createdAt,
		now,
		nil,
		nil,
		"",
		true,
	).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<time") {
		t.Fatalf("html = %q, want timestamp in header", html)
	}
	if !strings.Contains(html, "2h") {
		t.Fatalf("html = %q, want compact relative timestamp", html)
	}
}
