package post_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/post"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	g "maragu.dev/gomponents"
)

func TestPostSpine_WrapsItemsInListItems(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	var buf bytes.Buffer
	if err := post.PostSpine("spine-id", []g.Node{
		post.PostInThread(feedquery.PostView{ID: "a", Text: "one"}, now),
		post.PostInThread(feedquery.PostView{ID: "b", Text: "two"}, now),
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `id="spine-id"`) || !strings.Contains(html, `class="post-spine"`) {
		t.Fatalf("html = %q, want post-spine ul with id", html)
	}
	if strings.Count(html, "<li><article") != 2 {
		t.Fatalf("html = %q, want two spine list items", html)
	}
}
