package moderation

import "context"

type ProfileLabelView struct {
	Val           string
	Labeler       string
	Message       string
	LabelerDID      string
	LabelerHandle   string
	LabelerAvatar   string
	LabelerIsLabeler bool
}

// ProfileLabelsForDisplay returns known account/profile labels for the profile
// indicator, including labels whose preference is set to ignore.
func ProfileLabelsForDisplay(labels []Label, authorDID string, prefs Prefs) []ProfileLabelView {
	if len(labels) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	out := make([]ProfileLabelView, 0, len(labels))
	for _, label := range labels {
		if !profileLabelApplies(label, authorDID) {
			continue
		}
		def, ok := lookupLabelDefinition(label.Val)
		if !ok {
			continue
		}

		isSelf := label.Src == authorDID
		if !isSelf && !prefs.subscribedLabeler(label.Src) {
			continue
		}
		if isSelf && def.hasFlag(FlagNoSelf) {
			continue
		}

		dedupeKey := label.Val + "\x00" + label.Src
		if _, ok := seen[dedupeKey]; ok {
			continue
		}
		seen[dedupeKey] = struct{}{}
		out = append(out, ProfileLabelView{
			Val:        label.Val,
			Labeler:    LabelerDisplayName(label.Src, authorDID),
			Message:    def.Message,
			LabelerDID: label.Src,
		})
	}
	return out
}

func profileLabelApplies(label Label, authorDID string) bool {
	if label.URI == "" {
		return true
	}
	if authorDID != "" && label.URI == authorDID {
		return true
	}
	return IsProfileLabel(label)
}

// EvaluateProfileAvatar returns avatar blur state for a profile header.
func EvaluateProfileAvatar(ctx context.Context, provider PrefsProvider, authorDID string, labels []Label) UIResult {
	opts := Options{
		Prefs: provider.Prefs(ctx),
	}
	decision := ModeratePost(PostSubject{
		AuthorDID:    authorDID,
		AuthorLabels: labels,
	}, opts)
	return decision.UI(UIContextAvatar)
}
