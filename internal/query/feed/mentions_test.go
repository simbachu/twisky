package feed_test

import (
	"context"
	"errors"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/richtext"
)

type stubResolver struct {
	profiles []bluesky.Profile
	err      error
	calls    int
	last     []string
}

func (s *stubResolver) GetProfiles(_ context.Context, actors []string) ([]bluesky.Profile, error) {
	s.calls++
	s.last = append([]string(nil), actors...)
	if s.err != nil {
		return nil, s.err
	}
	return s.profiles, nil
}

func TestResolveMentionHandles_ResolvesDIDToHandle(t *testing.T) {
	t.Parallel()

	resolver := &stubResolver{
		profiles: []bluesky.Profile{{
			DID:    "did:plc:example",
			Handle: "dev.example",
		}},
	}
	view := feedquery.FeedView{
		Posts: []feedquery.PostView{{
			TextSegments: []richtext.Segment{{
				Kind:    richtext.Mention,
				Text:    "@dev.example",
				Mention: "did:plc:example",
			}},
		}},
	}

	got := feedquery.ResolveMentionHandles(context.Background(), resolver, view)

	if got.Posts[0].TextSegments[0].Mention != "dev.example" {
		t.Fatalf("mention = %q, want dev.example", got.Posts[0].TextSegments[0].Mention)
	}
	if resolver.calls != 1 {
		t.Fatalf("resolver calls = %d, want 1", resolver.calls)
	}
}

func TestResolveMentionHandles_LeavesUnresolvedDIDs(t *testing.T) {
	t.Parallel()

	resolver := &stubResolver{profiles: []bluesky.Profile{}}
	view := feedquery.FeedView{
		Posts: []feedquery.PostView{{
			TextSegments: []richtext.Segment{{
				Kind:    richtext.Mention,
				Text:    "@missing.example",
				Mention: "did:plc:missing",
			}},
		}},
	}

	got := feedquery.ResolveMentionHandles(context.Background(), resolver, view)

	if got.Posts[0].TextSegments[0].Mention != "did:plc:missing" {
		t.Fatalf("mention = %q, want did:plc:missing", got.Posts[0].TextSegments[0].Mention)
	}
}

func TestResolveMentionHandles_NoMentionsSkipsResolver(t *testing.T) {
	t.Parallel()

	resolver := &stubResolver{}
	view := feedquery.FeedView{
		Posts: []feedquery.PostView{{
			TextSegments: []richtext.Segment{{
				Kind: richtext.Plain,
				Text: "plain text",
			}},
		}},
	}

	got := feedquery.ResolveMentionHandles(context.Background(), resolver, view)

	if resolver.calls != 0 {
		t.Fatalf("resolver calls = %d, want 0", resolver.calls)
	}
	if got.Posts[0].TextSegments[0].Text != "plain text" {
		t.Fatalf("text = %q, want plain text", got.Posts[0].TextSegments[0].Text)
	}
}

func TestResolveMentionHandles_ResolverErrorKeepsDIDs(t *testing.T) {
	t.Parallel()

	resolver := &stubResolver{err: errors.New("network failure")}
	view := feedquery.FeedView{
		Posts: []feedquery.PostView{{
			TextSegments: []richtext.Segment{{
				Kind:    richtext.Mention,
				Text:    "@dev.example",
				Mention: "did:plc:example",
			}},
		}},
	}

	got := feedquery.ResolveMentionHandles(context.Background(), resolver, view)

	if got.Posts[0].TextSegments[0].Mention != "did:plc:example" {
		t.Fatalf("mention = %q, want did:plc:example", got.Posts[0].TextSegments[0].Mention)
	}
}

func TestResolveMentionHandles_ChunksRequests(t *testing.T) {
	t.Parallel()

	resolver := &stubResolver{
		profiles: []bluesky.Profile{{
			DID:    "did:plc:first",
			Handle: "first.example",
		}},
	}
	posts := make([]feedquery.PostView, 0, 26)
	for i := 0; i < 26; i++ {
		posts = append(posts, feedquery.PostView{
			TextSegments: []richtext.Segment{{
				Kind:    richtext.Mention,
				Text:    "@user",
				Mention: "did:plc:" + string(rune('a'+i)),
			}},
		})
	}
	view := feedquery.FeedView{Posts: posts}

	_ = feedquery.ResolveMentionHandles(context.Background(), resolver, view)

	if resolver.calls != 2 {
		t.Fatalf("resolver calls = %d, want 2", resolver.calls)
	}
	if len(resolver.last) != 1 {
		t.Fatalf("last chunk size = %d, want 1", len(resolver.last))
	}
}

func TestResolveMentionHandles_ResolvesNestedQuotedPostMentions(t *testing.T) {
	t.Parallel()

	resolver := &stubResolver{
		profiles: []bluesky.Profile{{
			DID:    "did:plc:quoted",
			Handle: "quoted.example",
		}},
	}
	quoted := feedquery.PostView{
		TextSegments: []richtext.Segment{{
			Kind:    richtext.Mention,
			Text:    "@quoted.example",
			Mention: "did:plc:quoted",
		}},
	}
	view := feedquery.FeedView{
		Posts: []feedquery.PostView{{
			Text:             "my take",
			QuotedPostMaybe:  &quoted,
		}},
	}

	got := feedquery.ResolveMentionHandles(context.Background(), resolver, view)

	if got.Posts[0].QuotedPostMaybe.TextSegments[0].Mention != "quoted.example" {
		t.Fatalf("quoted mention = %q, want quoted.example", got.Posts[0].QuotedPostMaybe.TextSegments[0].Mention)
	}
}

func TestResolveMentionHandles_ResolvesNestedReplyParentMentions(t *testing.T) {
	t.Parallel()

	resolver := &stubResolver{
		profiles: []bluesky.Profile{{
			DID:    "did:plc:parent",
			Handle: "parent.example",
		}},
	}
	parent := feedquery.PostView{
		TextSegments: []richtext.Segment{{
			Kind:    richtext.Mention,
			Text:    "@parent.example",
			Mention: "did:plc:parent",
		}},
	}
	view := feedquery.FeedView{
		Posts: []feedquery.PostView{{
			Text:             "a reply",
			ReplyParentMaybe: &parent,
		}},
	}

	got := feedquery.ResolveMentionHandles(context.Background(), resolver, view)

	if got.Posts[0].ReplyParentMaybe.TextSegments[0].Mention != "parent.example" {
		t.Fatalf("parent mention = %q, want parent.example", got.Posts[0].ReplyParentMaybe.TextSegments[0].Mention)
	}
}
