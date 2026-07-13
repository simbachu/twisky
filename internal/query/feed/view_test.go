package feed_test

import (
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/richtext"
)

func TestNewPostView_FallsBackToHandleWhenDisplayNameEmpty(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		Author: bluesky.Author{Handle: "dev.example"},
		Record: bluesky.PostRecord{
			Text:      "hello",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	})

	if view.AuthorDisplayName != "dev.example" {
		t.Fatalf("view.AuthorDisplayName = %q, want dev.example", view.AuthorDisplayName)
	}
}

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

func TestNewPostView_PopulatesVideosFromVideoEmbed(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		URI:    "at://did:plc:example/app.bsky.feed.post/video",
		Author: bluesky.Author{Handle: "dev.example"},
		Record: bluesky.PostRecord{
			Text:      "video post",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Embed: &bluesky.Embed{
			Type:         "app.bsky.embed.video#view",
			Playlist:     "https://video.example.com/playlist.m3u8",
			Thumbnail:    "https://video.example.com/thumb.jpg",
			Alt:          "a clip",
			Presentation: "default",
			AspectRatio: &bluesky.AspectRatio{
				Width:  1920,
				Height: 1080,
			},
		},
	})

	if len(view.Videos) != 1 {
		t.Fatalf("len(view.Videos) = %d, want 1", len(view.Videos))
	}
	video := view.Videos[0]
	if video.Playlist != "https://video.example.com/playlist.m3u8" {
		t.Fatalf("video.Playlist = %q, want playlist URL", video.Playlist)
	}
	if video.Thumbnail != "https://video.example.com/thumb.jpg" {
		t.Fatalf("video.Thumbnail = %q, want thumbnail URL", video.Thumbnail)
	}
	if video.Alt != "a clip" {
		t.Fatalf("video.Alt = %q, want a clip", video.Alt)
	}
	if video.Presentation != "default" {
		t.Fatalf("video.Presentation = %q, want default", video.Presentation)
	}
	if video.Width != 1920 || video.Height != 1080 {
		t.Fatalf("video dimensions = %dx%d, want 1920x1080", video.Width, video.Height)
	}
}

func TestNewPostView_PopulatesVideoFromRecordWithMediaEmbed(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		URI:    "at://did:plc:example/app.bsky.feed.post/qrt-video",
		Author: bluesky.Author{Handle: "dev.example"},
		Record: bluesky.PostRecord{
			Text:      "quote with video",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Embed: &bluesky.Embed{
			Type: "app.bsky.embed.recordWithMedia#view",
			Record: mustRawJSON(`{
				"$type": "app.bsky.embed.record#viewRecord",
				"uri": "at://did:plc:quoted/app.bsky.feed.post/with-image",
				"author": {"handle": "quoted.example"},
				"value": {"text": "has image", "createdAt": "2026-01-14T12:00:00.000Z"},
				"embeds": [{
					"$type": "app.bsky.embed.images#view",
					"images": [{
						"thumb": "https://example.com/thumb.jpg",
						"fullsize": "https://example.com/full.jpg",
						"alt": "quoted photo"
					}]
				}]
			}`),
			Media: &bluesky.Embed{
				Type:         "app.bsky.embed.video#view",
				Playlist:     "https://video.example.com/my.m3u8",
				Thumbnail:    "https://video.example.com/my-thumb.jpg",
				Alt:          "my clip",
				Presentation: "gif",
			},
		},
	})

	if len(view.Videos) != 1 || view.Videos[0].Alt != "my clip" {
		t.Fatalf("view.Videos = %#v, want my clip", view.Videos)
	}
	if len(view.Images) != 0 {
		t.Fatalf("len(view.Images) = %d, want 0 on top-level post", len(view.Images))
	}
	if view.QuotedPostMaybe == nil {
		t.Fatal("QuotedPostMaybe = nil, want quoted post")
	}
	if len(view.QuotedPostMaybe.Images) != 1 || view.QuotedPostMaybe.Images[0].Alt != "quoted photo" {
		t.Fatalf("quoted images = %#v, want quoted photo", view.QuotedPostMaybe.Images)
	}
}

func TestNewPostView_PopulatesLinkPreviewFromExternalEmbed(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		URI:    "at://did:plc:example/app.bsky.feed.post/link",
		Author: bluesky.Author{Handle: "dev.example"},
		Record: bluesky.PostRecord{
			Text:      "check this out",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Embed: &bluesky.Embed{
			Type: "app.bsky.embed.external#view",
			External: &bluesky.ExternalView{
				URI:         "https://example.com",
				Title:       "Example Site",
				Description: "An example website",
				Thumb:       "https://example.com/thumb.jpg",
			},
		},
	})

	if view.LinkPreviewMaybe == nil {
		t.Fatal("LinkPreviewMaybe = nil, want link preview")
	}
	if view.LinkPreviewMaybe.URI != "https://example.com" {
		t.Fatalf("LinkPreviewMaybe.URI = %q, want https://example.com", view.LinkPreviewMaybe.URI)
	}
	if view.LinkPreviewMaybe.Title != "Example Site" {
		t.Fatalf("LinkPreviewMaybe.Title = %q, want Example Site", view.LinkPreviewMaybe.Title)
	}
	if view.LinkPreviewMaybe.Description != "An example website" {
		t.Fatalf("LinkPreviewMaybe.Description = %q, want An example website", view.LinkPreviewMaybe.Description)
	}
	if view.LinkPreviewMaybe.Thumb != "https://example.com/thumb.jpg" {
		t.Fatalf("LinkPreviewMaybe.Thumb = %q, want thumb URL", view.LinkPreviewMaybe.Thumb)
	}
}

func TestNewPostView_PopulatesLinkPreviewFromRecordWithMediaExternal(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		URI:    "at://did:plc:example/app.bsky.feed.post/qrt-external",
		Author: bluesky.Author{Handle: "dev.example"},
		Record: bluesky.PostRecord{
			Text:      "quote with link card",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Embed: &bluesky.Embed{
			Type: "app.bsky.embed.recordWithMedia#view",
			Record: mustRawJSON(`{
				"$type": "app.bsky.embed.record#viewRecord",
				"uri": "at://did:plc:quoted/app.bsky.feed.post/original",
				"author": {"handle": "quoted.example"},
				"value": {"text": "original post", "createdAt": "2026-01-14T12:00:00.000Z"}
			}`),
			Media: &bluesky.Embed{
				Type: "app.bsky.embed.external#view",
				External: &bluesky.ExternalView{
					URI:         "https://example.com",
					Title:       "Example",
					Description: "Description",
				},
			},
		},
	})

	if view.LinkPreviewMaybe == nil {
		t.Fatal("LinkPreviewMaybe = nil, want link preview from record-with-media")
	}
	if view.LinkPreviewMaybe.Title != "Example" {
		t.Fatalf("LinkPreviewMaybe.Title = %q, want Example", view.LinkPreviewMaybe.Title)
	}
	if view.QuotedPostMaybe == nil {
		t.Fatal("QuotedPostMaybe = nil, want quoted post")
	}
}

func TestNewPostView_PopulatesLinkPreviewOnQuotedPost(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		URI:    "at://did:plc:example/app.bsky.feed.post/qrt",
		Author: bluesky.Author{Handle: "dev.example"},
		Record: bluesky.PostRecord{
			Text:      "sharing",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Embed: &bluesky.Embed{
			Type: "app.bsky.embed.record#view",
			Record: mustRawJSON(`{
				"$type": "app.bsky.embed.record#viewRecord",
				"uri": "at://did:plc:quoted/app.bsky.feed.post/with-link",
				"author": {"handle": "quoted.example"},
				"value": {"text": "read more", "createdAt": "2026-01-14T12:00:00.000Z"},
				"embeds": [{
					"$type": "app.bsky.embed.external#view",
					"external": {
						"uri": "https://example.com/article",
						"title": "Article Title",
						"description": "Article summary"
					}
				}]
			}`),
		},
	})

	if view.LinkPreviewMaybe != nil {
		t.Fatalf("LinkPreviewMaybe = %#v, want nil on parent post", view.LinkPreviewMaybe)
	}
	if view.QuotedPostMaybe == nil {
		t.Fatal("QuotedPostMaybe = nil, want quoted post")
	}
	if view.QuotedPostMaybe.LinkPreviewMaybe == nil {
		t.Fatal("quoted LinkPreviewMaybe = nil, want nested link preview")
	}
	if view.QuotedPostMaybe.LinkPreviewMaybe.Title != "Article Title" {
		t.Fatalf("quoted LinkPreviewMaybe.Title = %q, want Article Title", view.QuotedPostMaybe.LinkPreviewMaybe.Title)
	}
}
