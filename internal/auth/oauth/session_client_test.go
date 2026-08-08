package oauth_test

import (
	"context"
	"errors"
	"testing"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/bluesky"
)

func TestIsDuplicateLike(t *testing.T) {
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
		if got := authoauth.IsDuplicateLike(tc.err); got != tc.want {
			t.Fatalf("IsDuplicateLike(%v) = %v, want %v", tc.err, got, tc.want)
		}
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
