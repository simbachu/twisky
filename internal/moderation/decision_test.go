package moderation_test

import (
	"context"
	"testing"
	"time"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/moderation"
)

func TestModeratePost_PornWithAdultOff_FiltersAndBlursWithoutOverride(t *testing.T) {
	t.Parallel()

	listUI := moderate(t, moderation.Label{Val: "porn", Src: moderation.BlueskyModerationDID}, moderation.UIContextContentList)
	mediaUI := moderate(t, moderation.Label{Val: "porn", Src: moderation.BlueskyModerationDID}, moderation.UIContextContentMedia)

	if !listUI.Filter {
		t.Fatal("Filter = false, want true for porn with hide pref")
	}
	if !mediaUI.BlurMedia {
		t.Fatal("BlurMedia = false, want true for porn in contentMedia context")
	}
	if !mediaUI.NoOverride {
		t.Fatal("NoOverride = false, want true when adult content disabled")
	}
}

func TestModeratePost_Sexual_WarnsWithoutFiltering(t *testing.T) {
	t.Parallel()

	listUI := moderate(t, moderation.Label{Val: "sexual", Src: moderation.BlueskyModerationDID}, moderation.UIContextContentList)
	mediaUI := moderate(t, moderation.Label{Val: "sexual", Src: moderation.BlueskyModerationDID}, moderation.UIContextContentMedia)

	if listUI.Filter {
		t.Fatal("Filter = true, want false for sexual with warn pref")
	}
	if !mediaUI.BlurMedia {
		t.Fatal("BlurMedia = false, want true for sexual in contentMedia context")
	}
}

func TestModeratePost_NudityWithIgnorePref_TakesNoAction(t *testing.T) {
	t.Parallel()

	ui := moderate(t, moderation.Label{Val: "nudity", Src: moderation.BlueskyModerationDID}, moderation.UIContextContentList)

	if ui.Filter || ui.Blur || ui.BlurMedia || ui.Alert || ui.Inform {
		t.Fatalf("ui = %#v, want no moderation action for ignored nudity", ui)
	}
}

func TestModeratePost_NoUnauthenticated_FiltersForAnonymousViewer(t *testing.T) {
	t.Parallel()

	decision := moderation.ModeratePost(moderation.PostSubject{
		AuthorDID: "did:plc:author",
		Labels:    []moderation.Label{{Val: "!no-unauthenticated", Src: moderation.BlueskyModerationDID}},
	}, moderation.Options{Prefs: moderation.DefaultPrefs()})

	ui := decision.UI(moderation.UIContextContentList)
	if !ui.Filter {
		t.Fatal("Filter = false, want true for !no-unauthenticated")
	}
	if !ui.NoOverride {
		t.Fatal("NoOverride = false, want true for !no-unauthenticated")
	}
}

func TestModeratePost_SelfLabel_UsesGlobalPrefs(t *testing.T) {
	t.Parallel()

	decision := moderation.ModeratePost(moderation.PostSubject{
		AuthorDID: "did:plc:author",
		Labels:    []moderation.Label{{Val: "sexual", Src: "did:plc:author"}},
	}, moderation.Options{Prefs: moderation.DefaultPrefs()})

	listUI := decision.UI(moderation.UIContextContentList)
	mediaUI := decision.UI(moderation.UIContextContentMedia)
	if listUI.Filter {
		t.Fatal("Filter = true, want false for self-labeled sexual with warn pref")
	}
	if !mediaUI.BlurMedia {
		t.Fatal("BlurMedia = false, want true for self-labeled sexual")
	}
}

func TestEvaluatePost_UsesPrefsProvider(t *testing.T) {
	t.Parallel()

	provider := moderation.DefaultPrefsProvider{}
	ui := moderation.EvaluatePost(context.Background(), provider, bluesky.Post{
		Author: bluesky.Author{DID: "did:plc:author", Handle: "author.example"},
		Labels: []bluesky.Label{{Val: "porn", Src: moderation.BlueskyModerationDID}},
		Record: bluesky.PostRecord{
			Text:      "hello",
			CreatedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	}, moderation.UIContextContentList)
	if !ui.Filter {
		t.Fatal("Filter = false, want true from EvaluatePost")
	}
}

func moderate(t *testing.T, label moderation.Label, context moderation.UIContext) moderation.UIResult {
	t.Helper()

	decision := moderation.ModeratePost(moderation.PostSubject{
		AuthorDID: "did:plc:author",
		Labels:    []moderation.Label{label},
	}, moderation.Options{Prefs: moderation.DefaultPrefs()})
	return decision.UI(context)
}
