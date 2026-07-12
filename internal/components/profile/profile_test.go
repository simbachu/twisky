package profile_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/components/profile"
	profilequery "github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/richtext"
)

func TestProfile_RendersDescriptionFacets(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := profile.Profile(profilequery.ProfileView{
		Handle:      "dev.example",
		DisplayName: "Developer",
		Description: "@dev.example see https://example.com",
		DescriptionSegments: []richtext.Segment{
			{Kind: richtext.Mention, Text: "@dev.example", Mention: "dev.example"},
			{Kind: richtext.Plain, Text: " see "},
			{Kind: richtext.Link, Text: "https://example.com", URI: "https://example.com"},
		},
		Tab: profilequery.TabPosts,
	}, time.Now().UTC()).Render(&buf); err != nil {
		t.Fatalf("Render() err = %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `href="/dev.example"`) {
		t.Fatalf("html = %q, want mention link to /dev.example", html)
	}
	if !strings.Contains(html, `href="https://example.com"`) {
		t.Fatalf("html = %q, want external link to https://example.com", html)
	}
}
