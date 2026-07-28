package moderation_test

import (
	"context"
	"testing"

	"github.com/simbachu/twisky/internal/moderation"
)

func TestProfileLabelsForDisplay_IncludesIgnoredLabels(t *testing.T) {
	t.Parallel()

	labels := moderation.ProfileLabelsForDisplay(
		[]moderation.Label{{
			Val: "nudity",
			Src: moderation.BlueskyModerationDID,
			URI: "did:plc:author",
		}},
		"did:plc:author",
		moderation.DefaultPrefs(),
	)

	if len(labels) != 1 {
		t.Fatalf("len(labels) = %d, want 1", len(labels))
	}
	if labels[0].Labeler != moderation.BlueskyModerationName {
		t.Fatalf("Labeler = %q, want %q", labels[0].Labeler, moderation.BlueskyModerationName)
	}
	if labels[0].Message != "Non-sexual nudity" {
		t.Fatalf("Message = %q, want Non-sexual nudity", labels[0].Message)
	}
}

func TestProfileLabelsForDisplay_DeduplicatesByVal(t *testing.T) {
	t.Parallel()

	labels := moderation.ProfileLabelsForDisplay(
		[]moderation.Label{
			{Val: "sexual", Src: moderation.BlueskyModerationDID, URI: "did:plc:author"},
			{Val: "sexual", Src: moderation.BlueskyModerationDID, URI: "at://did:plc:author/app.bsky.actor.profile/self"},
		},
		"did:plc:author",
		moderation.DefaultPrefs(),
	)

	if len(labels) != 1 {
		t.Fatalf("len(labels) = %d, want 1", len(labels))
	}
}

func TestProfileLabelsForDisplay_ExcludesUnsubscribedLabeler(t *testing.T) {
	t.Parallel()

	labels := moderation.ProfileLabelsForDisplay(
		[]moderation.Label{{
			Val: "sexual",
			Src: "did:plc:unknown-labeler",
			URI: "did:plc:author",
		}},
		"did:plc:author",
		moderation.DefaultPrefs(),
	)

	if len(labels) != 0 {
		t.Fatalf("len(labels) = %d, want 0 for unsubscribed labeler", len(labels))
	}
}

func TestProfileLabelsForDisplay_ExcludesContentScopedLabels(t *testing.T) {
	t.Parallel()

	labels := moderation.ProfileLabelsForDisplay(
		[]moderation.Label{{
			Val: "porn",
			Src: moderation.BlueskyModerationDID,
			URI: "at://did:plc:author/app.bsky.feed.post/abc",
		}},
		"did:plc:author",
		moderation.DefaultPrefs(),
	)

	if len(labels) != 0 {
		t.Fatalf("len(labels) = %d, want 0 for content-scoped label", len(labels))
	}
}

func TestProfileLabelsForDisplay_IncludesProfileScopedLabels(t *testing.T) {
	t.Parallel()

	labels := moderation.ProfileLabelsForDisplay(
		[]moderation.Label{{
			Val: "sexual",
			Src: moderation.BlueskyModerationDID,
			URI: "at://did:plc:author/app.bsky.actor.profile/self",
		}},
		"did:plc:author",
		moderation.DefaultPrefs(),
	)

	if len(labels) != 1 {
		t.Fatalf("len(labels) = %d, want 1", len(labels))
	}
}

func TestEvaluateProfileAvatar_BlursForAccountPorn(t *testing.T) {
	t.Parallel()

	ui := moderation.EvaluateProfileAvatar(
		context.Background(),
		moderation.DefaultPrefsProvider{},
		"did:plc:author",
		[]moderation.Label{{
			Val: "porn",
			Src: moderation.BlueskyModerationDID,
			URI: "did:plc:author",
		}},
	)

	if !ui.BlurAvatar {
		t.Fatal("BlurAvatar = false, want true for account porn label")
	}
	if ui.PrimaryMessage() != "Adult content" {
		t.Fatalf("PrimaryMessage() = %q, want Adult content", ui.PrimaryMessage())
	}
}
