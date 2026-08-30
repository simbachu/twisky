package compose_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/compose"
	"github.com/simbachu/twisky/internal/components/ui"
)

func TestNewPostPage_RendersFieldAndBack(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := compose.NewPostPage("", "", "https://twisky.test", nil, ui.AccountMenuView{Enabled: true}).Render(&buf); err != nil {
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
