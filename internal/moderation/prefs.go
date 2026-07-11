package moderation

import "context"

type LabelerPrefs struct {
	DID    string
	Labels map[string]LabelPreference
}

type Prefs struct {
	AdultContentEnabled bool
	Labels              map[string]LabelPreference
	Labelers            []LabelerPrefs
}

type PrefsProvider interface {
	Prefs(ctx context.Context) Prefs
}

type DefaultPrefsProvider struct{}

func (DefaultPrefsProvider) Prefs(context.Context) Prefs {
	return DefaultPrefs()
}

func DefaultPrefs() Prefs {
	defaultLabels := map[string]LabelPreference{
		"porn":          LabelHide,
		"sexual":        LabelWarn,
		"nudity":        LabelIgnore,
		"graphic-media": LabelHide,
	}
	return Prefs{
		AdultContentEnabled: false,
		Labels:              defaultLabels,
		Labelers: []LabelerPrefs{{
			DID:    BlueskyModerationDID,
			Labels: defaultLabels,
		}},
	}
}

func (p Prefs) labelPreference(def LabelDefinition, labelerDID string, isSelf bool) LabelPreference {
	if !def.Configurable {
		if def.DefaultSetting != "" {
			return def.DefaultSetting
		}
		return LabelHide
	}
	if def.hasFlag(FlagAdult) && !p.AdultContentEnabled {
		return LabelHide
	}
	if !isSelf && labelerDID != "" {
		for _, labeler := range p.Labelers {
			if labeler.DID == labelerDID {
				if pref, ok := labeler.Labels[def.Identifier]; ok {
					return pref
				}
			}
		}
	}
	if pref, ok := p.Labels[def.Identifier]; ok {
		return pref
	}
	if def.DefaultSetting != "" {
		return def.DefaultSetting
	}
	return LabelIgnore
}

func (p Prefs) subscribedLabeler(did string) bool {
	for _, labeler := range p.Labelers {
		if labeler.DID == did {
			return true
		}
	}
	return false
}
