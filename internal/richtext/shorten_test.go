package richtext_test

import (
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/richtext"
)

func TestShortenLinks_StripsScheme(t *testing.T) {
	t.Parallel()

	text := "see https://example.com now"
	facets := richtext.DetectFacets(text)
	gotText, gotFacets := richtext.ShortenLinks(text, facets)

	if gotText != "see example.com now" {
		t.Fatalf("text = %q, want shortened", gotText)
	}
	if len(gotFacets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(gotFacets))
	}
	assertLinkFacet(t, gotFacets[0], 4, 15, "https://example.com")
	if gotText[gotFacets[0].Index.ByteStart:gotFacets[0].Index.ByteEnd] != "example.com" {
		t.Fatalf("display slice = %q", gotText[gotFacets[0].Index.ByteStart:gotFacets[0].Index.ByteEnd])
	}
}

func TestShortenLinks_TruncatesLongPath(t *testing.T) {
	t.Parallel()

	long := "https://example.com/very/long/path/here"
	text := "go " + long
	facets := richtext.DetectFacets(text)
	gotText, gotFacets := richtext.ShortenLinks(text, facets)

	wantDisplay := "example.com/very/long/pa..."
	wantText := "go " + wantDisplay
	if gotText != wantText {
		t.Fatalf("text = %q, want %q", gotText, wantText)
	}
	if len(gotFacets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(gotFacets))
	}
	assertLinkFacet(t, gotFacets[0], 3, 3+len(wantDisplay), long)
}

func TestShortenLinks_LeavesBareDomain(t *testing.T) {
	t.Parallel()

	text := "visit middle.com please"
	facets := richtext.DetectFacets(text)
	gotText, gotFacets := richtext.ShortenLinks(text, facets)

	if gotText != text {
		t.Fatalf("text = %q, want unchanged", gotText)
	}
	assertLinkFacet(t, gotFacets[0], 6, 16, "https://middle.com")
}

func TestShortenLinks_PreservesLaterFacets(t *testing.T) {
	t.Parallel()

	text := "https://example.com and #mtg"
	facets := richtext.DetectFacets(text)
	gotText, gotFacets := richtext.ShortenLinks(text, facets)

	if !strings.HasPrefix(gotText, "example.com") {
		t.Fatalf("text = %q", gotText)
	}
	if len(gotFacets) != 2 {
		t.Fatalf("len(facets) = %d, want 2", len(gotFacets))
	}
	assertLinkFacet(t, gotFacets[0], 0, 11, "https://example.com")
	assertTagFacet(t, gotFacets[1], 16, 20, "mtg")
}

func TestStripInvalidMentions_DropsHandles(t *testing.T) {
	t.Parallel()

	facets := []bluesky.Facet{
		{
			Index:    bluesky.FacetIndex{ByteStart: 0, ByteEnd: 9},
			Features: []bluesky.FacetFeature{{Type: "app.bsky.richtext.facet#mention", DID: "bsky.app"}},
		},
		{
			Index:    bluesky.FacetIndex{ByteStart: 10, ByteEnd: 14},
			Features: []bluesky.FacetFeature{{Type: "app.bsky.richtext.facet#tag", Tag: "mtg"}},
		},
		{
			Index:    bluesky.FacetIndex{ByteStart: 15, ByteEnd: 24},
			Features: []bluesky.FacetFeature{{Type: "app.bsky.richtext.facet#mention", DID: "did:plc:ok"}},
		},
	}
	got := richtext.StripInvalidMentions(facets)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	assertTagFacet(t, got[0], 10, 14, "mtg")
	assertMentionFacet(t, got[1], 15, 24, "did:plc:ok")
}
