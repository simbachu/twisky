package moderation_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/moderation"
)

func TestIsProfileLabel(t *testing.T) {
	t.Parallel()

	label := moderation.Label{
		Val: "porn",
		URI: "at://did:plc:author/app.bsky.actor.profile/self",
	}
	if !moderation.IsProfileLabel(label) {
		t.Fatal("IsProfileLabel() = false, want true for profile/self URI")
	}
}

func TestIsAccountAuthorLabel_ExcludesProfileLabels(t *testing.T) {
	t.Parallel()

	label := moderation.Label{
		Val: "porn",
		URI: "at://did:plc:author/app.bsky.actor.profile/self",
	}
	if moderation.IsAccountAuthorLabel(label) {
		t.Fatal("IsAccountAuthorLabel() = true, want false for profile-scoped label")
	}
}

func TestIsAccountAuthorLabel_IncludesRepoLevelLabels(t *testing.T) {
	t.Parallel()

	label := moderation.Label{
		Val: "porn",
		URI: "did:plc:author",
	}
	if !moderation.IsAccountAuthorLabel(label) {
		t.Fatal("IsAccountAuthorLabel() = false, want true for repo-level label")
	}
}

func TestIsAccountAuthorLabel_IncludesNoUnauthenticatedOnProfileURI(t *testing.T) {
	t.Parallel()

	label := moderation.Label{
		Val: "!no-unauthenticated",
		URI: "at://did:plc:author/app.bsky.actor.profile/self",
	}
	if !moderation.IsAccountAuthorLabel(label) {
		t.Fatal("IsAccountAuthorLabel() = false, want true for !no-unauthenticated on profile URI")
	}
}

func TestLabelAppliesToPost_MatchesPostURI(t *testing.T) {
	t.Parallel()

	postURI := "at://did:plc:author/app.bsky.feed.post/abc"
	label := moderation.Label{URI: postURI}

	if !moderation.LabelAppliesToPost(label, postURI, "did:plc:author") {
		t.Fatal("LabelAppliesToPost() = false, want true when URI matches post")
	}
}

func TestLabelAppliesToPost_MatchesAuthorDID(t *testing.T) {
	t.Parallel()

	label := moderation.Label{URI: "did:plc:author"}
	if !moderation.LabelAppliesToPost(label, "at://did:plc:author/app.bsky.feed.post/abc", "did:plc:author") {
		t.Fatal("LabelAppliesToPost() = false, want true for repo-level URI")
	}
}

func TestLabelAppliesToPost_EmptyURITrustsAPI(t *testing.T) {
	t.Parallel()

	if !moderation.LabelAppliesToPost(moderation.Label{}, "at://did:plc:author/app.bsky.feed.post/abc", "did:plc:author") {
		t.Fatal("LabelAppliesToPost() = false, want true for empty URI")
	}
}

func TestLabelAppliesToPost_RejectsMismatchedURI(t *testing.T) {
	t.Parallel()

	label := moderation.Label{URI: "at://did:plc:other/app.bsky.feed.post/other"}
	if moderation.LabelAppliesToPost(label, "at://did:plc:author/app.bsky.feed.post/abc", "did:plc:author") {
		t.Fatal("LabelAppliesToPost() = true, want false for mismatched post URI")
	}
}
