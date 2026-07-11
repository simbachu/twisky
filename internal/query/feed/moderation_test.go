package feed_test

import (
	"context"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/moderation"
)

func TestApplyModeration_DropsFilteredPostsFromFeed(t *testing.T) {
	t.Parallel()

	feed := feedquery.NewFeedView([]bluesky.Post{
		{
			URI:    "at://did:plc:author/app.bsky.feed.post/safe",
			Author: bluesky.Author{DID: "did:plc:author", Handle: "author.example"},
			Record: bluesky.PostRecord{
				Text:      "safe",
				CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			URI:    "at://did:plc:author/app.bsky.feed.post/adult",
			Author: bluesky.Author{DID: "did:plc:author", Handle: "author.example"},
			Record: bluesky.PostRecord{
				Text:      "adult",
				CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
			},
			Labels: []bluesky.Label{{Val: "porn", Src: moderation.BlueskyModerationDID}},
		},
	}, "")

	moderated := feedquery.ApplyModeration(context.Background(), moderation.DefaultPrefsProvider{}, feed, moderation.UIContextContentList)
	if len(moderated.Posts) != 1 {
		t.Fatalf("len(posts) = %d, want 1 after filtering porn", len(moderated.Posts))
	}
	if moderated.Posts[0].Text != "safe" {
		t.Fatalf("remaining post text = %q, want safe", moderated.Posts[0].Text)
	}
}

func TestApplyModeration_BlursSexualMedia(t *testing.T) {
	t.Parallel()

	feed := feedquery.NewFeedView([]bluesky.Post{{
		URI:    "at://did:plc:author/app.bsky.feed.post/suggestive",
		Author: bluesky.Author{DID: "did:plc:author", Handle: "author.example"},
		Record: bluesky.PostRecord{
			Text:      "hello",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Labels: []bluesky.Label{{Val: "sexual", Src: moderation.BlueskyModerationDID}},
	}}, "")

	moderated := feedquery.ApplyModeration(context.Background(), moderation.DefaultPrefsProvider{}, feed, moderation.UIContextContentList)
	if len(moderated.Posts) != 1 {
		t.Fatalf("len(posts) = %d, want 1", len(moderated.Posts))
	}
	if !moderated.Posts[0].Moderation.BlurMedia {
		t.Fatal("BlurMedia = false, want true for sexual label")
	}
}

func TestApplyModeration_OmitsFilteredQuotedPost(t *testing.T) {
	t.Parallel()

	quoted := bluesky.Post{
		URI:    "at://did:plc:quoted/app.bsky.feed.post/quoted",
		Author: bluesky.Author{DID: "did:plc:quoted", Handle: "quoted.example"},
		Record: bluesky.PostRecord{
			Text:      "quoted",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Labels: []bluesky.Label{{Val: "porn", Src: moderation.BlueskyModerationDID}},
	}
	mainPost := bluesky.Post{
		URI:    "at://did:plc:author/app.bsky.feed.post/main",
		Author: bluesky.Author{DID: "did:plc:author", Handle: "author.example"},
		Record: bluesky.PostRecord{
			Text:      "quote",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		Embed: &bluesky.Embed{
			Type: "app.bsky.embed.record#view",
		},
	}
	mainPost.Embed.Record = mustRecordEmbed(t, quoted)

	view := feedquery.NewPostView(mainPost)
	moderated := feedquery.ApplyModeration(context.Background(), moderation.DefaultPrefsProvider{}, feedquery.FeedView{Posts: []feedquery.PostView{view}}, moderation.UIContextContentList)
	if len(moderated.Posts) != 1 {
		t.Fatalf("len(posts) = %d, want 1", len(moderated.Posts))
	}
	if moderated.Posts[0].QuotedPostMaybe != nil {
		t.Fatal("QuotedPostMaybe = non-nil, want filtered quote omitted")
	}
}

func mustRecordEmbed(t *testing.T, post bluesky.Post) []byte {
	t.Helper()
	raw := `{
		"$type": "app.bsky.embed.record#viewRecord",
		"uri": "` + post.URI + `",
		"author": {
			"did": "` + post.Author.DID + `",
			"handle": "` + post.Author.Handle + `"
		},
		"value": {
			"text": "` + post.Record.Text + `",
			"createdAt": "2026-01-15T12:00:00.000Z"
		},
		"labels": [{"val": "porn", "src": "` + moderation.BlueskyModerationDID + `"}]
	}`
	return []byte(raw)
}
