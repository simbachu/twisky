package feed_test

import (
	"testing"

	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func TestNewPostsSince_EmptySinceID(t *testing.T) {
	t.Parallel()

	posts := []feedquery.PostView{
		{ID: "a"},
		{ID: "b"},
	}

	got := feedquery.NewPostsSince(posts, "")

	if got != nil {
		t.Fatalf("NewPostsSince() = %v, want nil", got)
	}
}

func TestNewPostsSince_MatchInMiddle(t *testing.T) {
	t.Parallel()

	posts := []feedquery.PostView{
		{ID: "new-1"},
		{ID: "new-2"},
		{ID: "known"},
		{ID: "older"},
	}

	got := feedquery.NewPostsSince(posts, "known")

	if len(got) != 2 {
		t.Fatalf("len(NewPostsSince()) = %d, want 2", len(got))
	}
	if got[0].ID != "new-1" || got[1].ID != "new-2" {
		t.Fatalf("NewPostsSince() = %v, want [new-1 new-2]", got)
	}
}

func TestNewPostsSince_MatchAtIndexZero(t *testing.T) {
	t.Parallel()

	posts := []feedquery.PostView{
		{ID: "top"},
		{ID: "older"},
	}

	got := feedquery.NewPostsSince(posts, "top")

	if got != nil {
		t.Fatalf("NewPostsSince() = %v, want nil", got)
	}
}

func TestNewPostsSince_NoMatch(t *testing.T) {
	t.Parallel()

	posts := []feedquery.PostView{
		{ID: "a"},
		{ID: "b"},
	}

	got := feedquery.NewPostsSince(posts, "missing")

	if len(got) != 2 {
		t.Fatalf("len(NewPostsSince()) = %d, want 2", len(got))
	}
}
