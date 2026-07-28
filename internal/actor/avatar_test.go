package actor_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/moderation"
)

func TestAvatarFallbackEmoji_LabelerShield(t *testing.T) {
	t.Parallel()

	if got := actor.AvatarFallbackEmoji("moderation.bsky.app", moderation.BlueskyModerationDID, false); got != actor.AvatarEmojiShield {
		t.Fatalf("AvatarFallbackEmoji() = %q, want shield", got)
	}
}

func TestAvatarFallbackEmoji_BskyButterfly(t *testing.T) {
	t.Parallel()

	if got := actor.AvatarFallbackEmoji("bsky.app", "did:plc:bsky", false); got != actor.AvatarEmojiButterfly {
		t.Fatalf("AvatarFallbackEmoji() = %q, want butterfly", got)
	}
}

func TestAvatarFallbackEmoji_EmptyForOtherAccounts(t *testing.T) {
	t.Parallel()

	if got := actor.AvatarFallbackEmoji("alice.bsky.social", "did:plc:alice", false); got != "" {
		t.Fatalf("AvatarFallbackEmoji() = %q, want empty", got)
	}
}

func TestIsLabelerAccount_FromAPIFlag(t *testing.T) {
	t.Parallel()

	if !actor.IsLabelerAccount("", "", true) {
		t.Fatal("IsLabelerAccount() = false, want true for API labeler flag")
	}
}
