package oauth_test

import (
	"context"
	"errors"
	"testing"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/bluesky"
)

func TestIsDuplicateRecord(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("pds down"), false},
		{errors.New("InvalidRequest: Record already exists"), true},
		{errors.New("DuplicateCreate"), true},
	}
	for _, tc := range cases {
		if got := authoauth.IsDuplicateRecord(tc.err); got != tc.want {
			t.Fatalf("IsDuplicateRecord(%v) = %v, want %v", tc.err, got, tc.want)
		}
		if got := authoauth.IsDuplicateLike(tc.err); got != tc.want {
			t.Fatalf("IsDuplicateLike(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}

func TestIsRecordNotFound(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("pds down"), false},
		{errors.New("RecordNotFound: could not locate record"), true},
		{errors.New("not found"), true},
	}
	for _, tc := range cases {
		if got := authoauth.IsRecordNotFound(tc.err); got != tc.want {
			t.Fatalf("IsRecordNotFound(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}

func TestSessionClient_CreateRepost_NilClient(t *testing.T) {
	t.Parallel()

	client := authoauth.NewSessionClient(nil)
	_, err := client.CreateRepost(context.Background(), "at://did:plc:example/app.bsky.feed.post/abc", "bafycid")
	if err == nil {
		t.Fatal("CreateRepost() err = nil, want error")
	}
}

func TestSessionClient_DeleteLike_NilClient(t *testing.T) {
	t.Parallel()

	client := authoauth.NewSessionClient(nil)
	err := client.DeleteLike(context.Background(), "at://did:plc:me/app.bsky.feed.like/like1")
	if err == nil {
		t.Fatal("DeleteLike() err = nil, want error")
	}
}

func TestSessionClient_DeleteRepost_NilClient(t *testing.T) {
	t.Parallel()

	client := authoauth.NewSessionClient(nil)
	err := client.DeleteRepost(context.Background(), "at://did:plc:me/app.bsky.feed.repost/r1")
	if err == nil {
		t.Fatal("DeleteRepost() err = nil, want error")
	}
}

func TestSessionClient_GetTimeline_NilClient(t *testing.T) {
	t.Parallel()

	client := authoauth.NewSessionClient(nil)
	_, err := client.GetTimeline(context.Background(), bluesky.TimelineRequest{Limit: 20})
	if err == nil {
		t.Fatal("GetTimeline() err = nil, want error")
	}
}

func TestSessionClient_GetProfiles_NilClient(t *testing.T) {
	t.Parallel()

	client := authoauth.NewSessionClient(nil)
	_, err := client.GetProfiles(context.Background(), []string{"did:plc:example"})
	if err == nil {
		t.Fatal("GetProfiles() err = nil, want error")
	}
}

func TestSessionClient_GetSavedFeeds_NilClient(t *testing.T) {
	t.Parallel()

	client := authoauth.NewSessionClient(nil)

	_, err := client.GetSavedFeeds(context.Background())

	if err == nil {
		t.Fatal("GetSavedFeeds() err = nil, want error")
	}
}

func TestSessionClient_GetFeedGenerators_NilClient(t *testing.T) {
	t.Parallel()

	client := authoauth.NewSessionClient(nil)

	_, err := client.GetFeedGenerators(context.Background(), []string{"at://did:plc:example/app.bsky.feed.generator/for-you"})

	if err == nil {
		t.Fatal("GetFeedGenerators() err = nil, want error")
	}
}

func TestSessionClient_GetFeed_NilClient(t *testing.T) {
	t.Parallel()

	client := authoauth.NewSessionClient(nil)

	_, err := client.GetFeed(context.Background(), bluesky.FeedRequest{
		URI: "at://did:plc:example/app.bsky.feed.generator/for-you",
	})

	if err == nil {
		t.Fatal("GetFeed() err = nil, want error")
	}
}
