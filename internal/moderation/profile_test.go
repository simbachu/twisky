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
	if labels[0].LabelerDID != moderation.BlueskyModerationDID {
		t.Fatalf("LabelerDID = %q, want %q", labels[0].LabelerDID, moderation.BlueskyModerationDID)
	}
	if labels[0].Labeler != moderation.BlueskyModerationName {
		t.Fatalf("Labeler = %q, want %q", labels[0].Labeler, moderation.BlueskyModerationName)
	}
	if labels[0].Message != "Non-sexual nudity" {
		t.Fatalf("Message = %q, want Non-sexual nudity", labels[0].Message)
	}
}

func TestProfileLabelsForDisplay_DeduplicatesByValAndSrc(t *testing.T) {
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
		t.Fatalf("len(labels) = %d, want 1 for duplicate val+src", len(labels))
	}
}

func TestProfileLabelsForDisplay_AllowsSameValFromDifferentLabelers(t *testing.T) {
	t.Parallel()

	otherLabeler := "did:plc:other-labeler"
	prefs := moderation.DefaultPrefs()
	prefs.Labelers = append(prefs.Labelers, moderation.LabelerPrefs{DID: otherLabeler})

	labels := moderation.ProfileLabelsForDisplay(
		[]moderation.Label{
			{Val: "sexual", Src: moderation.BlueskyModerationDID, URI: "did:plc:author"},
			{Val: "sexual", Src: otherLabeler, URI: "did:plc:author"},
		},
		"did:plc:author",
		prefs,
	)

	if len(labels) != 2 {
		t.Fatalf("len(labels) = %d, want 2 for same val from different labelers", len(labels))
	}
}

func TestLabelerProfileSlug(t *testing.T) {
	t.Parallel()

	if got := moderation.LabelerProfileSlug(moderation.BlueskyModerationDID); got != "moderation.bsky.app" {
		t.Fatalf("LabelerProfileSlug() = %q, want moderation.bsky.app", got)
	}
	if got := moderation.LabelerProfileSlug("did:plc:custom"); got != "did:plc:custom" {
		t.Fatalf("LabelerProfileSlug() = %q, want did passthrough", got)
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
