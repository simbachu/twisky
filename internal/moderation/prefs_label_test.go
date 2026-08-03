package moderation

import (
	"testing"
)

func TestLabelPreference_Cascade(t *testing.T) {
	t.Parallel()

	porn, ok := lookupLabelDefinition("porn")
	if !ok {
		t.Fatal("missing porn definition")
	}
	hide, ok := lookupLabelDefinition("!hide")
	if !ok {
		t.Fatal("missing !hide definition")
	}
	nudity, ok := lookupLabelDefinition("nudity")
	if !ok {
		t.Fatal("missing nudity definition")
	}

	tests := []struct {
		name       string
		prefs      Prefs
		def        LabelDefinition
		labelerDID string
		isSelf     bool
		want       LabelPreference
	}{
		{
			name: "nonconfigurable uses default",
			prefs: Prefs{
				Labels: map[string]LabelPreference{"!hide": LabelIgnore},
			},
			def:  hide,
			want: LabelHide,
		},
		{
			name: "adult disabled forces hide",
			prefs: Prefs{
				AdultContentEnabled: false,
				Labels:              map[string]LabelPreference{"porn": LabelWarn},
			},
			def:        porn,
			labelerDID: BlueskyModerationDID,
			want:       LabelHide,
		},
		{
			name: "labeler-specific override",
			prefs: Prefs{
				AdultContentEnabled: true,
				Labels:              map[string]LabelPreference{"porn": LabelHide},
				Labelers: []LabelerPrefs{{
					DID:    BlueskyModerationDID,
					Labels: map[string]LabelPreference{"porn": LabelWarn},
				}},
			},
			def:        porn,
			labelerDID: BlueskyModerationDID,
			want:       LabelWarn,
		},
		{
			name: "self skips labeler map and uses global",
			prefs: Prefs{
				AdultContentEnabled: true,
				Labels:              map[string]LabelPreference{"porn": LabelWarn},
				Labelers: []LabelerPrefs{{
					DID:    BlueskyModerationDID,
					Labels: map[string]LabelPreference{"porn": LabelHide},
				}},
			},
			def:        porn,
			labelerDID: BlueskyModerationDID,
			isSelf:     true,
			want:       LabelWarn,
		},
		{
			name: "falls back to definition default",
			prefs: Prefs{
				AdultContentEnabled: true,
			},
			def:  nudity,
			want: LabelIgnore, // nudity DefaultSetting is ignore via built-in
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.prefs.labelPreference(tt.def, tt.labelerDID, tt.isSelf)
			if got != tt.want {
				t.Fatalf("labelPreference() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddLabel_PriorityAndGates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		authorDID  string
		label      Label
		opts       Options
		wantCauses int
		wantPref   LabelPreference
		wantPrio   int
	}{
		{
			name:      "unsubscribed labeler ignored",
			authorDID: "did:plc:author",
			label:     Label{Val: "porn", Src: "did:plc:stranger"},
			opts: Options{
				Prefs: Prefs{
					AdultContentEnabled: true,
					Labelers:            []LabelerPrefs{{DID: BlueskyModerationDID}},
				},
			},
			wantCauses: 0,
		},
		{
			name:      "self label accepted without subscription",
			authorDID: "did:plc:author",
			label:     Label{Val: "sexual", Src: "did:plc:author"},
			opts: Options{
				Prefs: Prefs{
					Labels: map[string]LabelPreference{"sexual": LabelWarn},
				},
			},
			wantCauses: 1,
			wantPref:   LabelWarn,
			wantPrio:   7, // contentMedia blur
		},
		{
			name:      "hide pref gets priority 2",
			authorDID: "did:plc:author",
			label:     Label{Val: "porn", Src: BlueskyModerationDID},
			opts: Options{
				Prefs: Prefs{
					AdultContentEnabled: true,
					Labels:              map[string]LabelPreference{"porn": LabelHide},
					Labelers: []LabelerPrefs{{
						DID:    BlueskyModerationDID,
						Labels: map[string]LabelPreference{"porn": LabelHide},
					}},
				},
			},
			wantCauses: 1,
			wantPref:   LabelHide,
			wantPrio:   2,
		},
		{
			name:      "adult disabled sets no-override priority 1",
			authorDID: "did:plc:author",
			label:     Label{Val: "porn", Src: BlueskyModerationDID},
			opts: Options{
				Prefs: Prefs{
					AdultContentEnabled: false,
					Labels:              map[string]LabelPreference{"porn": LabelHide},
					Labelers: []LabelerPrefs{{
						DID:    BlueskyModerationDID,
						Labels: map[string]LabelPreference{"porn": LabelHide},
					}},
				},
			},
			wantCauses: 1,
			wantPref:   LabelHide,
			wantPrio:   1,
		},
		{
			name:      "ignore pref drops label",
			authorDID: "did:plc:author",
			label:     Label{Val: "nudity", Src: BlueskyModerationDID},
			opts: Options{
				Prefs: Prefs{
					Labels:   map[string]LabelPreference{"nudity": LabelIgnore},
					Labelers: []LabelerPrefs{{DID: BlueskyModerationDID}},
				},
			},
			wantCauses: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := &Decision{authorDID: tt.authorDID}
			d.addLabel(LabelTargetContent, tt.label, tt.opts)
			if len(d.causes) != tt.wantCauses {
				t.Fatalf("causes = %d, want %d (%#v)", len(d.causes), tt.wantCauses, d.causes)
			}
			if tt.wantCauses == 0 {
				return
			}
			cause, ok := d.causes[0].(LabelCause)
			if !ok {
				t.Fatalf("cause type %T", d.causes[0])
			}
			if cause.Setting != tt.wantPref {
				t.Fatalf("Setting = %q, want %q", cause.Setting, tt.wantPref)
			}
			if cause.Priority != tt.wantPrio {
				t.Fatalf("Priority = %d, want %d", cause.Priority, tt.wantPrio)
			}
		})
	}
}
