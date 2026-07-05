package feed_test

import (
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/richtext"
)

func TestNewPostView_UsesPostRkeyAsID(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		URI:    "at://did:plc:example/app.bsky.feed.post/abc123",
		Author: bluesky.Author{Handle: "dev.example"},
		Record: bluesky.PostRecord{
			Text:      "hello",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	})

	if view.ID != "abc123" {
		t.Fatalf("view.ID = %q, want abc123", view.ID)
	}
}

func TestNewPostView_PopulatesTextSegmentsFromFacets(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		Author: bluesky.Author{Handle: "dev.example"},
		Record: bluesky.PostRecord{
			Text:      "hello #golang",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
			Facets: []bluesky.Facet{{
				Index: bluesky.FacetIndex{ByteStart: 6, ByteEnd: 13},
				Features: []bluesky.FacetFeature{{
					Type: "app.bsky.richtext.facet#tag",
					Tag:  "golang",
				}},
			}},
		},
	})

	if len(view.TextSegments) != 2 {
		t.Fatalf("len(view.TextSegments) = %d, want 2", len(view.TextSegments))
	}
	if view.TextSegments[0].Kind != richtext.Plain || view.TextSegments[0].Text != "hello " {
		t.Fatalf("first segment = %#v, want plain hello ", view.TextSegments[0])
	}
	if view.TextSegments[1].Kind != richtext.Tag || view.TextSegments[1].Tag != "golang" {
		t.Fatalf("second segment = %#v, want tag golang", view.TextSegments[1])
	}
}

func TestNewPostView_PopulatesMentionAndLinkSegmentsFromFacets(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		Author: bluesky.Author{Handle: "dev.example"},
		Record: bluesky.PostRecord{
			Text:      "@dev.example see https://example.com",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
			Facets: []bluesky.Facet{
				{
					Index: bluesky.FacetIndex{ByteStart: 0, ByteEnd: 12},
					Features: []bluesky.FacetFeature{{
						Type: "app.bsky.richtext.facet#mention",
						DID:  "did:plc:example",
					}},
				},
				{
					Index: bluesky.FacetIndex{ByteStart: 17, ByteEnd: 36},
					Features: []bluesky.FacetFeature{{
						Type: "app.bsky.richtext.facet#link",
						URI:  "https://example.com",
					}},
				},
			},
		},
	})

	if len(view.TextSegments) != 3 {
		t.Fatalf("len(view.TextSegments) = %d, want 3", len(view.TextSegments))
	}
	if view.TextSegments[0].Kind != richtext.Mention {
		t.Fatalf("first segment kind = %v, want Mention", view.TextSegments[0].Kind)
	}
	if view.TextSegments[2].Kind != richtext.Link || view.TextSegments[2].URI != "https://example.com" {
		t.Fatalf("third segment = %#v, want link https://example.com", view.TextSegments[2])
	}
}

func TestNewPostView_PopulatesImagesFromGalleryEmbed(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		URI:    "at://did:plc:example/app.bsky.feed.post/gallery",
		Author: bluesky.Author{Handle: "bsky.app", DisplayName: "Bluesky"},
		Record: bluesky.PostRecord{
			Text:      "gallery post",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Embed: &bluesky.Embed{
			Type: "app.bsky.embed.gallery#view",
			Items: []bluesky.EmbedImage{
				{
					Thumbnail: "https://example.com/thumb1.jpg",
					Fullsize:  "https://example.com/full1.jpg",
					AspectRatio: &bluesky.AspectRatio{
						Width:  1000,
						Height: 800,
					},
				},
				{
					Thumbnail: "https://example.com/thumb2.jpg",
					Fullsize:  "https://example.com/full2.jpg",
					Alt:       "second",
					AspectRatio: &bluesky.AspectRatio{
						Width:  1200,
						Height: 900,
					},
				},
			},
		},
	})

	if len(view.Images) != 2 {
		t.Fatalf("len(view.Images) = %d, want 2", len(view.Images))
	}
	if view.Images[0].Thumb != "https://example.com/thumb1.jpg" {
		t.Fatalf("view.Images[0].Thumb = %q, want https://example.com/thumb1.jpg", view.Images[0].Thumb)
	}
	if view.Images[0].Fullsize != "https://example.com/full1.jpg" {
		t.Fatalf("view.Images[0].Fullsize = %q, want https://example.com/full1.jpg", view.Images[0].Fullsize)
	}
	if view.Images[0].Width != 1000 || view.Images[0].Height != 800 {
		t.Fatalf("view.Images[0] dimensions = %dx%d, want 1000x800", view.Images[0].Width, view.Images[0].Height)
	}
	if view.Images[1].Alt != "second" {
		t.Fatalf("view.Images[1].Alt = %q, want second", view.Images[1].Alt)
	}
}
