package richtext_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/richtext"
)

func TestDetectFacets_HTTPSLink(t *testing.T) {
	t.Parallel()

	facets := richtext.DetectFacets("see https://example.com/page now")
	if len(facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(facets))
	}
	assertLinkFacet(t, facets[0], 4, 28, "https://example.com/page")
}

func TestDetectFacets_TrailingPunctuationStrippedFromSpan(t *testing.T) {
	t.Parallel()

	facets := richtext.DetectFacets("see https://example.com.")
	if len(facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(facets))
	}
	assertLinkFacet(t, facets[0], 4, 23, "https://example.com")
}

func TestDetectFacets_BareDomainAddsHTTPS(t *testing.T) {
	t.Parallel()

	facets := richtext.DetectFacets("visit middle.com today")
	if len(facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(facets))
	}
	assertLinkFacet(t, facets[0], 6, 16, "https://middle.com")
}

func TestDetectFacets_BareDomainInParens(t *testing.T) {
	t.Parallel()

	facets := richtext.DetectFacets("(see bsky.app)")
	if len(facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(facets))
	}
	assertLinkFacet(t, facets[0], 5, 13, "https://bsky.app")
}

func TestDetectFacets_RejectsVersionLookingDomain(t *testing.T) {
	t.Parallel()

	if facets := richtext.DetectFacets("v1.127 is live!"); facets != nil {
		t.Fatalf("facets = %#v, want nil", facets)
	}
}

func TestDetectFacets_RejectsFileExtension(t *testing.T) {
	t.Parallel()

	if facets := richtext.DetectFacets("open photo.jpg please"); facets != nil {
		t.Fatalf("facets = %#v, want nil", facets)
	}
}

func TestDetectFacets_Mention(t *testing.T) {
	t.Parallel()

	facets := richtext.DetectFacets("hello @bsky.app world")
	if len(facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(facets))
	}
	assertMentionFacet(t, facets[0], 6, 15, "bsky.app")
}

func TestDetectFacets_Tag(t *testing.T) {
	t.Parallel()

	facets := richtext.DetectFacets("hello #golang world")
	if len(facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(facets))
	}
	assertTagFacet(t, facets[0], 6, 13, "golang")
}

func TestDetectFacets_Cashtag(t *testing.T) {
	t.Parallel()

	facets := richtext.DetectFacets("buy $AAPL now")
	if len(facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(facets))
	}
	assertTagFacet(t, facets[0], 4, 9, "$AAPL")
}

func TestDetectFacets_UTF8ByteOffsets(t *testing.T) {
	t.Parallel()

	text := "café https://example.com"
	facets := richtext.DetectFacets(text)
	if len(facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(facets))
	}
	assertLinkFacet(t, facets[0], 6, 25, "https://example.com")
}

func TestDetectFacets_Mixed(t *testing.T) {
	t.Parallel()

	facets := richtext.DetectFacets("@bsky.app #mtg https://example.com")
	if len(facets) != 3 {
		t.Fatalf("len(facets) = %d, want 3", len(facets))
	}
	assertMentionFacet(t, facets[0], 0, 9, "bsky.app")
	assertTagFacet(t, facets[1], 10, 14, "mtg")
	assertLinkFacet(t, facets[2], 15, 34, "https://example.com")
}

func TestDetectSegments_UsesDetectFacets(t *testing.T) {
	t.Parallel()

	segments := richtext.DetectSegments("see middle.com")
	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Plain, text: "see "})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Link, text: "middle.com", uri: "https://middle.com"})
}

func assertLinkFacet(t *testing.T, facet bluesky.Facet, start, end int, uri string) {
	t.Helper()
	if facet.Index.ByteStart != start || facet.Index.ByteEnd != end {
		t.Fatalf("index = {%d,%d}, want {%d,%d}", facet.Index.ByteStart, facet.Index.ByteEnd, start, end)
	}
	if len(facet.Features) != 1 || facet.Features[0].Type != "app.bsky.richtext.facet#link" {
		t.Fatalf("features = %#v, want link", facet.Features)
	}
	if facet.Features[0].URI != uri {
		t.Fatalf("uri = %q, want %q", facet.Features[0].URI, uri)
	}
}

func assertMentionFacet(t *testing.T, facet bluesky.Facet, start, end int, did string) {
	t.Helper()
	if facet.Index.ByteStart != start || facet.Index.ByteEnd != end {
		t.Fatalf("index = {%d,%d}, want {%d,%d}", facet.Index.ByteStart, facet.Index.ByteEnd, start, end)
	}
	if len(facet.Features) != 1 || facet.Features[0].Type != "app.bsky.richtext.facet#mention" {
		t.Fatalf("features = %#v, want mention", facet.Features)
	}
	if facet.Features[0].DID != did {
		t.Fatalf("did = %q, want %q", facet.Features[0].DID, did)
	}
}

func assertTagFacet(t *testing.T, facet bluesky.Facet, start, end int, tag string) {
	t.Helper()
	if facet.Index.ByteStart != start || facet.Index.ByteEnd != end {
		t.Fatalf("index = {%d,%d}, want {%d,%d}", facet.Index.ByteStart, facet.Index.ByteEnd, start, end)
	}
	if len(facet.Features) != 1 || facet.Features[0].Type != "app.bsky.richtext.facet#tag" {
		t.Fatalf("features = %#v, want tag", facet.Features)
	}
	if facet.Features[0].Tag != tag {
		t.Fatalf("tag = %q, want %q", facet.Features[0].Tag, tag)
	}
}
