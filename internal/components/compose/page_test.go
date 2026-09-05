package compose_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/compose"
	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func TestNewPostPage_RendersFieldAndBack(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := compose.NewPostPage(compose.PageView{
		PublicBaseURL: "https://twisky.test",
		Accounts:      ui.AccountMenuView{Enabled: true},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`data-nav-back`,
		`>New post<`,
		`action="/my/posts"`,
		`id="new-post-text"`,
		`rel="canonical" href="https://twisky.test/my/posts/new"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %q", html, want)
		}
	}
}

func TestNewPostPage_RendersReplyParent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := compose.NewPostPage(compose.PageView{
		PublicBaseURL: "https://twisky.test",
		ParentURI:     "at://did:plc:example/app.bsky.feed.post/parent1",
		Parent:        Article(g.Attr("class", "post inset-post"), P(g.Text("parent body"))),
		Accounts:      ui.AccountMenuView{Enabled: true},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`>Reply<`,
		`class="post inset-post"`,
		"parent body",
		`name="parent"`,
		`value="at://did:plc:example/app.bsky.feed.post/parent1"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html = %q, want %q", html, want)
		}
	}
}
