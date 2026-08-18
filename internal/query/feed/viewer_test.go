package feed_test

import (
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestApplyViewerLikes_SetsLikedFromPosts(t *testing.T) {
	t.Parallel()

	views := []feedquery.PostView{
		feedquery.NewPostView(bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/a",
			CID:    "cida",
			Author: bluesky.Author{Handle: "a.example"},
			Record: bluesky.PostRecord{Text: "a", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		}),
		feedquery.NewPostView(bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/b",
			CID:    "cidb",
			Author: bluesky.Author{Handle: "b.example"},
			Record: bluesky.PostRecord{Text: "b", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		}),
	}

	feedquery.ApplyViewerLikes(views, []bluesky.Post{
		{
			URI:    "at://did:plc:example/app.bsky.feed.post/a",
			Viewer: &bluesky.PostViewer{Like: "at://did:plc:me/app.bsky.feed.like/1"},
		},
	})

	if !views[0].Liked {
		t.Fatal("views[0].Liked = false, want true")
	}
	if views[1].Liked {
		t.Fatal("views[1].Liked = true, want false")
	}
}

func TestApplyViewerLikes_SetsRepostedFromPosts(t *testing.T) {
	t.Parallel()

	views := []feedquery.PostView{
		feedquery.NewPostView(bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/a",
			CID:    "cida",
			Author: bluesky.Author{Handle: "a.example"},
			Record: bluesky.PostRecord{Text: "a", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		}),
		feedquery.NewPostView(bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/b",
			CID:    "cidb",
			Author: bluesky.Author{Handle: "b.example"},
			Record: bluesky.PostRecord{Text: "b", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		}),
	}

	feedquery.ApplyViewerLikes(views, []bluesky.Post{
		{
			URI:    "at://did:plc:example/app.bsky.feed.post/a",
			Viewer: &bluesky.PostViewer{Repost: "at://did:plc:me/app.bsky.feed.repost/1"},
		},
	})

	if !views[0].Reposted {
		t.Fatal("views[0].Reposted = false, want true")
	}
	if views[1].Reposted {
		t.Fatal("views[1].Reposted = true, want false")
	}
}

func TestCollectPostURIs(t *testing.T) {
	t.Parallel()

	views := []feedquery.PostView{
		feedquery.NewPostView(bluesky.Post{
			URI:    "at://did:plc:example/app.bsky.feed.post/a",
			Author: bluesky.Author{Handle: "a.example"},
			Record: bluesky.PostRecord{Text: "a", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		}),
	}
	uris := feedquery.CollectPostURIs(views)
	if len(uris) != 1 || uris[0] != "at://did:plc:example/app.bsky.feed.post/a" {
		t.Fatalf("uris = %#v", uris)
	}
}
