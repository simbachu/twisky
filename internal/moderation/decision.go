package moderation

type Label struct {
	Val string
	Src string
}

type Options struct {
	UserDID string
	Prefs   Prefs
}

type LabelCause struct {
	Label       Label
	Definition  LabelDefinition
	Target      LabelTarget
	Setting     LabelPreference
	Behavior    Behavior
	NoOverride  bool
	Priority    int
	Message     string
}

type UIResult struct {
	Filter     bool
	Blur       bool
	BlurMedia  bool
	Alert      bool
	Inform     bool
	NoOverride bool
	Alerts     []LabelCause
	Informs    []LabelCause
	Blurs      []LabelCause
}

type Decision struct {
	authorDID string
	causes    []any
}

func (d *Decision) addLabel(target LabelTarget, label Label, opts Options) {
	def, ok := lookupLabelDefinition(label.Val)
	if !ok {
		return
	}

	isSelf := label.Src == d.authorDID
	if !isSelf && !opts.Prefs.subscribedLabeler(label.Src) {
		return
	}
	if isSelf && def.hasFlag(FlagNoSelf) {
		return
	}

	pref := opts.Prefs.labelPreference(def, label.Src, isSelf)
	if pref == LabelIgnore {
		return
	}
	if def.hasFlag(FlagUnauthed) && opts.UserDID != "" {
		return
	}

	behavior := def.Behaviors[target]
	noOverride := def.hasFlag(FlagNoOverride)
	if def.hasFlag(FlagAdult) && !opts.Prefs.AdultContentEnabled {
		noOverride = true
	}

	priority := 8
	if noOverride {
		priority = 1
	} else if pref == LabelHide {
		priority = 2
	} else if behavior.action(UIContextContentView) == BehaviorBlur {
		priority = 5
	} else if behavior.action(UIContextContentList) == BehaviorBlur || behavior.action(UIContextContentMedia) == BehaviorBlur {
		priority = 7
	}

	d.causes = append(d.causes, LabelCause{
		Label:      label,
		Definition: def,
		Target:     target,
		Setting:    pref,
		Behavior:   behavior,
		NoOverride: noOverride,
		Priority:   priority,
		Message:    def.Message,
	})
}

func (d Decision) UI(context UIContext) UIResult {
	result := UIResult{}
	for _, cause := range d.causes {
		labelCause, ok := cause.(LabelCause)
		if !ok {
			continue
		}

		if (context == UIContextContentList || context == UIContextContentView) &&
			(labelCause.Target == LabelTargetAccount || labelCause.Target == LabelTargetContent) &&
			labelCause.Setting == LabelHide {
			result.Filter = true
		}

		action := labelCause.Behavior.action(context)
		switch action {
		case BehaviorBlur:
			if context == UIContextContentMedia {
				result.BlurMedia = true
			} else {
				result.Blur = true
			}
			result.Blurs = append(result.Blurs, labelCause)
			if labelCause.NoOverride {
				result.NoOverride = true
			}
		case BehaviorAlert:
			result.Alert = true
			result.Alerts = append(result.Alerts, labelCause)
		case BehaviorInform:
			result.Inform = true
			result.Informs = append(result.Informs, labelCause)
		}
	}
	return result
}

func (r UIResult) PrimaryMessage() string {
	ui := r
	if len(ui.Alerts) > 0 {
		return ui.Alerts[0].Message
	}
	if len(ui.Informs) > 0 {
		return ui.Informs[0].Message
	}
	if len(ui.Blurs) > 0 {
		return ui.Blurs[0].Message
	}
	return ""
}
