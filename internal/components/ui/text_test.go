package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
	"github.com/simbachu/twisky/internal/richtext"
)

func TestRichText_RendersFacetSpanClasses(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := ui.RichText([]richtext.Segment{
		{Kind: richtext.Mention, Text: "@dev.example", Mention: "dev.example"},
		{Kind: richtext.Plain, Text: " see "},
		{Kind: richtext.Tag, Text: "#golang", Tag: "golang"},
		{Kind: richtext.Plain, Text: " "},
		{Kind: richtext.Link, Text: "https://example.com", URI: "https://example.com"},
	}).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	for _, class := range []string{`class="facet-mention"`, `class="facet-tag"`, `class="facet-link"`} {
		if !strings.Contains(html, class) {
			t.Fatalf("html = %q, want %s", html, class)
		}
	}
	if !strings.Contains(html, "🡕") {
		t.Fatalf("html = %q, want external link indicator", html)
	}
}
