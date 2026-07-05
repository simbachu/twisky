package richtext_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/richtext"
)

func TestBuildSegments_FromFacets(t *testing.T) {
	t.Parallel()

	text := "hello #mtg world"
	segments := richtext.BuildSegments(text, []bluesky.Facet{{
		Index: bluesky.FacetIndex{ByteStart: 6, ByteEnd: 10},
		Features: []bluesky.FacetFeature{{
			Type: "app.bsky.richtext.facet#tag",
			Tag:  "mtg",
		}},
	}})

	if len(segments) != 3 {
		t.Fatalf("len(segments) = %d, want 3", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Plain, text: "hello "})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Tag, text: "#mtg", tag: "mtg"})
	assertSegment(t, segments[2], segmentExpect{kind: richtext.Plain, text: " world"})
}

func TestBuildSegments_FromFacetsMention(t *testing.T) {
	t.Parallel()

	text := "hi @bsky.app there"
	segments := richtext.BuildSegments(text, []bluesky.Facet{{
		Index: bluesky.FacetIndex{ByteStart: 3, ByteEnd: 12},
		Features: []bluesky.FacetFeature{{
			Type: "app.bsky.richtext.facet#mention",
			DID:  "did:plc:example",
		}},
	}})

	if len(segments) != 3 {
		t.Fatalf("len(segments) = %d, want 3", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Plain, text: "hi "})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Mention, text: "@bsky.app", mention: "did:plc:example"})
	assertSegment(t, segments[2], segmentExpect{kind: richtext.Plain, text: " there"})
}

func TestBuildSegments_FromFacetsLink(t *testing.T) {
	t.Parallel()

	text := "see https://example.com now"
	segments := richtext.BuildSegments(text, []bluesky.Facet{{
		Index: bluesky.FacetIndex{ByteStart: 4, ByteEnd: 23},
		Features: []bluesky.FacetFeature{{
			Type: "app.bsky.richtext.facet#link",
			URI:  "https://example.com",
		}},
	}})

	if len(segments) != 3 {
		t.Fatalf("len(segments) = %d, want 3", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Plain, text: "see "})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Link, text: "https://example.com", uri: "https://example.com"})
	assertSegment(t, segments[2], segmentExpect{kind: richtext.Plain, text: " now"})
}

func TestBuildSegments_FromFacetsMixed(t *testing.T) {
	t.Parallel()

	text := "@bsky.app #mtg https://example.com"
	segments := richtext.BuildSegments(text, []bluesky.Facet{
		{
			Index: bluesky.FacetIndex{ByteStart: 0, ByteEnd: 9},
			Features: []bluesky.FacetFeature{{
				Type: "app.bsky.richtext.facet#mention",
				DID:  "did:plc:example",
			}},
		},
		{
			Index: bluesky.FacetIndex{ByteStart: 10, ByteEnd: 14},
			Features: []bluesky.FacetFeature{{
				Type: "app.bsky.richtext.facet#tag",
				Tag:  "mtg",
			}},
		},
		{
			Index: bluesky.FacetIndex{ByteStart: 15, ByteEnd: 34},
			Features: []bluesky.FacetFeature{{
				Type: "app.bsky.richtext.facet#link",
				URI:  "https://example.com",
			}},
		},
	})

	if len(segments) != 5 {
		t.Fatalf("len(segments) = %d, want 5", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Mention, text: "@bsky.app", mention: "did:plc:example"})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Plain, text: " "})
	assertSegment(t, segments[2], segmentExpect{kind: richtext.Tag, text: "#mtg", tag: "mtg"})
	assertSegment(t, segments[3], segmentExpect{kind: richtext.Plain, text: " "})
	assertSegment(t, segments[4], segmentExpect{kind: richtext.Link, text: "https://example.com", uri: "https://example.com"})
}

func TestBuildSegments_FromFacetsMultipleTags(t *testing.T) {
	t.Parallel()

	text := "#edh #mtg"
	segments := richtext.BuildSegments(text, []bluesky.Facet{
		{
			Index: bluesky.FacetIndex{ByteStart: 0, ByteEnd: 4},
			Features: []bluesky.FacetFeature{{
				Type: "app.bsky.richtext.facet#tag",
				Tag:  "edh",
			}},
		},
		{
			Index: bluesky.FacetIndex{ByteStart: 5, ByteEnd: 9},
			Features: []bluesky.FacetFeature{{
				Type: "app.bsky.richtext.facet#tag",
				Tag:  "mtg",
			}},
		},
	})

	if len(segments) != 3 {
		t.Fatalf("len(segments) = %d, want 3", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Tag, text: "#edh", tag: "edh"})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Plain, text: " "})
	assertSegment(t, segments[2], segmentExpect{kind: richtext.Tag, text: "#mtg", tag: "mtg"})
}

func TestBuildSegments_FromFacetsUTF8(t *testing.T) {
	t.Parallel()

	text := "café #mtg"
	segments := richtext.BuildSegments(text, []bluesky.Facet{{
		Index: bluesky.FacetIndex{ByteStart: 6, ByteEnd: 10},
		Features: []bluesky.FacetFeature{{
			Type: "app.bsky.richtext.facet#tag",
			Tag:  "mtg",
		}},
	}})

	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Plain, text: "café "})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Tag, text: "#mtg", tag: "mtg"})
}

func TestBuildSegments_RegexFallbackTag(t *testing.T) {
	t.Parallel()

	segments := richtext.BuildSegments("hello #golang world", nil)

	if len(segments) != 3 {
		t.Fatalf("len(segments) = %d, want 3", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Plain, text: "hello "})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Tag, text: "#golang", tag: "golang"})
	assertSegment(t, segments[2], segmentExpect{kind: richtext.Plain, text: " world"})
}

func TestBuildSegments_RegexFallbackMention(t *testing.T) {
	t.Parallel()

	segments := richtext.BuildSegments("hello @bsky.app world", nil)

	if len(segments) != 3 {
		t.Fatalf("len(segments) = %d, want 3", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Plain, text: "hello "})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Mention, text: "@bsky.app", mention: "bsky.app"})
	assertSegment(t, segments[2], segmentExpect{kind: richtext.Plain, text: " world"})
}

func TestBuildSegments_RegexFallbackLink(t *testing.T) {
	t.Parallel()

	segments := richtext.BuildSegments("see https://example.com/page", nil)

	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Plain, text: "see "})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Link, text: "https://example.com/page", uri: "https://example.com/page"})
}

func TestBuildSegments_RegexFallbackTrailingPunctuation(t *testing.T) {
	t.Parallel()

	segments := richtext.BuildSegments("nice #mtg!", nil)

	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Plain, text: "nice "})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Tag, text: "#mtg!", tag: "mtg"})
}

func TestBuildSegments_MentionFacetsWithoutTagFacetsDoesNotFallback(t *testing.T) {
	t.Parallel()

	segments := richtext.BuildSegments("hi @bsky.app", []bluesky.Facet{{
		Index: bluesky.FacetIndex{ByteStart: 3, ByteEnd: 12},
		Features: []bluesky.FacetFeature{{
			Type: "app.bsky.richtext.facet#mention",
			DID:  "did:plc:example",
		}},
	}})

	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	assertSegment(t, segments[0], segmentExpect{kind: richtext.Plain, text: "hi "})
	assertSegment(t, segments[1], segmentExpect{kind: richtext.Mention, text: "@bsky.app", mention: "did:plc:example"})
}

func TestBuildSegments_NoLinks(t *testing.T) {
	t.Parallel()

	if segments := richtext.BuildSegments("plain text", nil); segments != nil {
		t.Fatalf("segments = %#v, want nil", segments)
	}
}

type segmentExpect struct {
	kind    richtext.SegmentKind
	text    string
	tag     string
	mention string
	uri     string
}

func assertSegment(t *testing.T, segment richtext.Segment, want segmentExpect) {
	t.Helper()
	if segment.Kind != want.kind {
		t.Fatalf("segment.Kind = %v, want %v", segment.Kind, want.kind)
	}
	if segment.Text != want.text {
		t.Fatalf("segment.Text = %q, want %q", segment.Text, want.text)
	}
	if segment.Tag != want.tag {
		t.Fatalf("segment.Tag = %q, want %q", segment.Tag, want.tag)
	}
	if segment.Mention != want.mention {
		t.Fatalf("segment.Mention = %q, want %q", segment.Mention, want.mention)
	}
	if segment.URI != want.uri {
		t.Fatalf("segment.URI = %q, want %q", segment.URI, want.uri)
	}
}
