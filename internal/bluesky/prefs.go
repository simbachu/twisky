package bluesky

import (
	"encoding/json"
)

const (
	PrefTypeAdultContent = "app.bsky.actor.defs#adultContentPref"
	PrefTypeContentLabel = "app.bsky.actor.defs#contentLabelPref"
	PrefTypeSavedFeedsV2 = "app.bsky.actor.defs#savedFeedsPrefV2"
	PrefTypeSavedFeeds   = "app.bsky.actor.defs#savedFeedsPref"
	PrefTypeThreadView   = "app.bsky.actor.defs#threadViewPref"
)

const (
	LabelIgnore LabelVisibility = "ignore"
	LabelWarn   LabelVisibility = "warn"
	LabelHide   LabelVisibility = "hide"
)

const (
	ThreadSortOldest    = "oldest"
	ThreadSortNewest    = "newest"
	ThreadSortMostLikes = "most-likes"
	ThreadSortHotness   = "hotness"
)

// ContentFilterLabels is the built-in global content-label order shown in settings.
var ContentFilterLabels = []string{"porn", "sexual", "nudity", "graphic-media"}

// LabelVisibility is a Bluesky content-label preference (show maps to ignore).
type LabelVisibility string

type ThreadViewPref struct {
	Sort                    string
	PrioritizeFollowedUsers bool
}

// Preferences is a parsed app.bsky.actor.getPreferences payload.
// Raw is the full array and must be sent unchanged except for merged types.
type Preferences struct {
	Raw                 []json.RawMessage
	AdultContentEnabled bool
	Labels              map[string]LabelVisibility
	SavedFeeds          []SavedFeed
	ThreadView          ThreadViewPref
}

func defaultContentLabels() map[string]LabelVisibility {
	return map[string]LabelVisibility{
		"porn":          LabelHide,
		"sexual":        LabelWarn,
		"nudity":        LabelIgnore,
		"graphic-media": LabelHide,
	}
}

func parseVisibility(raw string) LabelVisibility {
	switch raw {
	case string(LabelHide):
		return LabelHide
	case string(LabelWarn):
		return LabelWarn
	default:
		return LabelIgnore
	}
}

func prefType(raw json.RawMessage) (string, error) {
	var header struct {
		Type string `json:"$type"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return "", err
	}
	return header.Type, nil
}

func isContentFilterLabel(identifier string) bool {
	for _, id := range ContentFilterLabels {
		if id == identifier {
			return true
		}
	}
	return false
}

func parseThreadSort(raw string) string {
	switch raw {
	case ThreadSortOldest, ThreadSortNewest, ThreadSortMostLikes, ThreadSortHotness:
		return raw
	default:
		return ThreadSortHotness
	}
}

// ParsePreferences extracts the preference types Twisky surfaces, keeping Raw intact.
func ParsePreferences(raw []json.RawMessage) (Preferences, error) {
	prefs := Preferences{
		Raw:        append([]json.RawMessage(nil), raw...),
		Labels:     defaultContentLabels(),
		ThreadView: ThreadViewPref{Sort: ThreadSortHotness},
	}
	var legacyFeeds []SavedFeed
	for _, item := range raw {
		kind, err := prefType(item)
		if err != nil {
			return Preferences{}, err
		}
		switch kind {
		case PrefTypeAdultContent:
			var pref struct {
				Enabled bool `json:"enabled"`
			}
			if err := json.Unmarshal(item, &pref); err != nil {
				return Preferences{}, err
			}
			prefs.AdultContentEnabled = pref.Enabled
		case PrefTypeContentLabel:
			var pref struct {
				LabelerDID string `json:"labelerDid"`
				Label      string `json:"label"`
				Visibility string `json:"visibility"`
			}
			if err := json.Unmarshal(item, &pref); err != nil {
				return Preferences{}, err
			}
			if pref.LabelerDID != "" || !isContentFilterLabel(pref.Label) {
				continue
			}
			prefs.Labels[pref.Label] = parseVisibility(pref.Visibility)
		case PrefTypeSavedFeedsV2:
			var pref struct {
				Items []SavedFeed `json:"items"`
			}
			if err := json.Unmarshal(item, &pref); err != nil {
				return Preferences{}, err
			}
			prefs.SavedFeeds = pref.Items
		case PrefTypeSavedFeeds:
			var pref struct {
				Pinned []string `json:"pinned"`
			}
			if err := json.Unmarshal(item, &pref); err != nil {
				return Preferences{}, err
			}
			legacyFeeds = make([]SavedFeed, 0, len(pref.Pinned))
			for _, uri := range pref.Pinned {
				legacyFeeds = append(legacyFeeds, SavedFeed{
					ID:     uri,
					Pinned: true,
					Type:   "feed",
					URI:    uri,
				})
			}
		case PrefTypeThreadView:
			var pref struct {
				Sort                    string `json:"sort"`
				PrioritizeFollowedUsers bool   `json:"prioritizeFollowedUsers"`
			}
			if err := json.Unmarshal(item, &pref); err != nil {
				return Preferences{}, err
			}
			prefs.ThreadView = ThreadViewPref{
				Sort:                    parseThreadSort(pref.Sort),
				PrioritizeFollowedUsers: pref.PrioritizeFollowedUsers,
			}
		}
	}
	if prefs.SavedFeeds == nil {
		prefs.SavedFeeds = legacyFeeds
	}
	return prefs, nil
}

func marshalPref(value any) (json.RawMessage, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func dropPrefType(raw []json.RawMessage, typeNSID string) []json.RawMessage {
	out := make([]json.RawMessage, 0, len(raw))
	for _, item := range raw {
		kind, err := prefType(item)
		if err != nil {
			out = append(out, item)
			continue
		}
		if kind == typeNSID {
			continue
		}
		out = append(out, item)
	}
	return out
}

func dropGlobalContentLabels(raw []json.RawMessage) []json.RawMessage {
	out := make([]json.RawMessage, 0, len(raw))
	for _, item := range raw {
		kind, err := prefType(item)
		if err != nil {
			continue
		}
		if kind != PrefTypeContentLabel {
			out = append(out, item)
			continue
		}
		var pref struct {
			LabelerDID string `json:"labelerDid"`
			Label      string `json:"label"`
		}
		if err := json.Unmarshal(item, &pref); err != nil {
			out = append(out, item)
			continue
		}
		if pref.LabelerDID == "" && isContentFilterLabel(pref.Label) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func (p Preferences) withRaw(raw []json.RawMessage) (Preferences, error) {
	return ParsePreferences(raw)
}

// WithAdultContent replaces adultContentPref and keeps every other preference type.
func (p Preferences) WithAdultContent(enabled bool) (Preferences, error) {
	replacement, err := marshalPref(map[string]any{
		"$type":   PrefTypeAdultContent,
		"enabled": enabled,
	})
	if err != nil {
		return Preferences{}, err
	}
	return p.withRaw(append(dropPrefType(p.Raw, PrefTypeAdultContent), replacement))
}

// WithContentLabels replaces global built-in contentLabelPref entries only.
func (p Preferences) WithContentLabels(labels map[string]LabelVisibility) (Preferences, error) {
	raw := dropGlobalContentLabels(p.Raw)
	for _, identifier := range ContentFilterLabels {
		visibility := labels[identifier]
		if visibility == "" {
			visibility = defaultContentLabels()[identifier]
		}
		replacement, err := marshalPref(map[string]any{
			"$type":      PrefTypeContentLabel,
			"label":      identifier,
			"visibility": string(visibility),
		})
		if err != nil {
			return Preferences{}, err
		}
		raw = append(raw, replacement)
	}
	return p.withRaw(raw)
}

// WithThreadView replaces threadViewPref and keeps every other preference type.
func (p Preferences) WithThreadView(view ThreadViewPref) (Preferences, error) {
	replacement, err := marshalPref(map[string]any{
		"$type":                   PrefTypeThreadView,
		"sort":                    parseThreadSort(view.Sort),
		"prioritizeFollowedUsers": view.PrioritizeFollowedUsers,
	})
	if err != nil {
		return Preferences{}, err
	}
	return p.withRaw(append(dropPrefType(p.Raw, PrefTypeThreadView), replacement))
}
