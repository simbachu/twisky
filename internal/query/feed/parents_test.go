package feed_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestNewPostView_PopulatesQuotedPostMaybe(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		URI:    "at://did:plc:example/app.bsky.feed.post/qrt",
		Author: bluesky.Author{Handle: "dev.example", DisplayName: "Dev"},
		Record: bluesky.PostRecord{
			Text:      "my take",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Embed: &bluesky.Embed{
			Type: "app.bsky.embed.record#view",
			Record: mustRawJSON(`{
				"$type": "app.bsky.embed.record#viewRecord",
				"uri": "at://did:plc:quoted/app.bsky.feed.post/original",
				"author": {"handle": "quoted.example", "displayName": "Quoted"},
				"value": {
					"text": "original post",
					"createdAt": "2026-01-14T12:00:00.000Z"
				}
			}`),
		},
	})

	if view.QuotedPostMaybe == nil {
		t.Fatal("QuotedPostMaybe = nil, want quoted post")
	}
	if view.QuotedPostMaybe.ID != "original" {
		t.Fatalf("QuotedPostMaybe.ID = %q, want original", view.QuotedPostMaybe.ID)
	}
	if view.QuotedPostMaybe.AuthorHandle != "quoted.example" {
		t.Fatalf("QuotedPostMaybe.AuthorHandle = %q, want quoted.example", view.QuotedPostMaybe.AuthorHandle)
	}
	if view.QuotedPostMaybe.Text != "original post" {
		t.Fatalf("QuotedPostMaybe.Text = %q, want original post", view.QuotedPostMaybe.Text)
	}
	if view.QuotedPostMaybe.QuotedPostMaybe != nil {
		t.Fatal("QuotedPostMaybe should not nest further quoted posts")
	}
}

func TestNewPostView_PopulatesQuotedPostImages(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostView(bluesky.Post{
		URI:    "at://did:plc:example/app.bsky.feed.post/qrt",
		Author: bluesky.Author{Handle: "dev.example"},
		Record: bluesky.PostRecord{
			Text:      "quote",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Embed: &bluesky.Embed{
			Type: "app.bsky.embed.record#view",
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
		},
	})

	if view.QuotedPostMaybe == nil {
		t.Fatal("QuotedPostMaybe = nil, want quoted post")
	}
	if len(view.QuotedPostMaybe.Images) != 1 {
		t.Fatalf("len(QuotedPostMaybe.Images) = %d, want 1", len(view.QuotedPostMaybe.Images))
	}
	if view.QuotedPostMaybe.Images[0].Alt != "quoted photo" {
		t.Fatalf("quoted image alt = %q, want quoted photo", view.QuotedPostMaybe.Images[0].Alt)
	}
}

func TestNewPostViewFromFeedItem_PopulatesReplyParentMaybe(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostViewFromFeedItem(bluesky.FeedItem{
		Post: bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/reply",
			Author: bluesky.Author{Handle: "dev.example"},
			Record: bluesky.PostRecord{
				Text:      "a reply",
				CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
		Reply: &bluesky.ReplyContext{
			Parent: &bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/parent",
				Author: bluesky.Author{Handle: "other.example", DisplayName: "Other"},
				Record: bluesky.PostRecord{
					Text:      "parent post",
					CreatedAt: time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC),
				},
			},
		},
	})

	if view.ReplyParentMaybe == nil {
		t.Fatal("ReplyParentMaybe = nil, want parent post")
	}
	if view.ReplyParentMaybe.ID != "parent" {
		t.Fatalf("ReplyParentMaybe.ID = %q, want parent", view.ReplyParentMaybe.ID)
	}
	if view.ReplyParentMaybe.Text != "parent post" {
		t.Fatalf("ReplyParentMaybe.Text = %q, want parent post", view.ReplyParentMaybe.Text)
	}
}

func TestNewPostViewFromFeedItem_PopulatesRepostedByMaybe(t *testing.T) {
	t.Parallel()

	view := feedquery.NewPostViewFromFeedItem(bluesky.FeedItem{
		Post: bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/original",
			Author: bluesky.Author{Handle: "original.example", DisplayName: "Original"},
			Record: bluesky.PostRecord{
				Text:      "original post",
				CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
		Reason: &bluesky.FeedReason{
			RepostedBy: bluesky.Author{Handle: "reposter.example", DisplayName: "Reposter"},
		},
	})

	if view.RepostedByMaybe == nil {
		t.Fatal("RepostedByMaybe = nil, want reposter")
	}
	if view.RepostedByMaybe.Handle != "reposter.example" {
		t.Fatalf("RepostedByMaybe.Handle = %q, want reposter.example", view.RepostedByMaybe.Handle)
	}
	if view.RepostedByMaybe.DisplayName != "Reposter" {
		t.Fatalf("RepostedByMaybe.DisplayName = %q, want Reposter", view.RepostedByMaybe.DisplayName)
	}
}

func TestEnrichReplyParents_HydratesParentFromGetPosts(t *testing.T) {
	t.Parallel()

	parentURI := "at://did:plc:example/app.bsky.feed.post/parent"
	feed := feedquery.FeedView{
		Posts: []feedquery.PostView{
			feedquery.NewPostView(bluesky.Post{
				URI:    "at://did:plc:example/app.bsky.feed.post/reply",
				Author: bluesky.Author{Handle: "dev.example"},
				Record: bluesky.PostRecord{
					Text:      "a reply",
					CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
					Reply: &bluesky.RecordReplyRef{
						Root:   bluesky.StrongRef{URI: "at://did:plc:example/app.bsky.feed.post/root"},
						Parent: bluesky.StrongRef{URI: parentURI},
					},
				},
			}),
		},
	}

	fetcher := stubPostFetcher{posts: []bluesky.Post{
		{
			URI:    parentURI,
			Author: bluesky.Author{Handle: "other.example", DisplayName: "Other"},
			Record: bluesky.PostRecord{
				Text:      "parent post",
				CreatedAt: time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC),
			},
		},
	}}

	enriched, err := feedquery.EnrichReplyParents(context.Background(), fetcher, feed)
	if err != nil {
		t.Fatalf("EnrichReplyParents() err = %v", err)
	}
	if enriched.Posts[0].ReplyParentMaybe == nil {
		t.Fatal("ReplyParentMaybe = nil, want hydrated parent")
	}
	if enriched.Posts[0].ReplyParentMaybe.Text != "parent post" {
		t.Fatalf("parent text = %q, want parent post", enriched.Posts[0].ReplyParentMaybe.Text)
	}
}

func TestEnrichReplyParents_SkipsWhenParentAlreadyPresent(t *testing.T) {
	t.Parallel()

	existingParent := feedquery.PostView{
		ID:           "parent",
		AuthorHandle: "other.example",
		Text:         "already here",
	}
	feed := feedquery.FeedView{
		Posts: []feedquery.PostView{{
			ID:               "reply",
			AuthorHandle:     "dev.example",
			Text:             "a reply",
			ReplyParentMaybe: &existingParent,
		}},
	}

	fetcher := stubPostFetcher{err: errors.New("should not be called")}

	enriched, err := feedquery.EnrichReplyParents(context.Background(), fetcher, feed)
	if err != nil {
		t.Fatalf("EnrichReplyParents() err = %v", err)
	}
	if enriched.Posts[0].ReplyParentMaybe.Text != "already here" {
		t.Fatalf("parent text = %q, want already here", enriched.Posts[0].ReplyParentMaybe.Text)
	}
}

type stubPostFetcher struct {
	posts []bluesky.Post
	err   error
}

func (s stubPostFetcher) GetPosts(context.Context, []string) ([]bluesky.Post, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.posts, nil
}

func mustRawJSON(value string) []byte {
	return []byte(value)
}
