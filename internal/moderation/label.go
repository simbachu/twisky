package moderation

const BlueskyModerationDID = "did:plc:ar7c4by46qjdydhdevvrndac"

type LabelTarget string

const (
	LabelTargetAccount LabelTarget = "account"
	LabelTargetProfile LabelTarget = "profile"
	LabelTargetContent LabelTarget = "content"
)

type LabelPreference string

const (
	LabelIgnore LabelPreference = "ignore"
	LabelWarn   LabelPreference = "warn"
	LabelHide   LabelPreference = "hide"
)

type BehaviorAction string

const (
	BehaviorBlur   BehaviorAction = "blur"
	BehaviorAlert  BehaviorAction = "alert"
	BehaviorInform BehaviorAction = "inform"
)

type UIContext string

const (
	UIContextContentList  UIContext = "contentList"
	UIContextContentView  UIContext = "contentView"
	UIContextContentMedia UIContext = "contentMedia"
)

type LabelDefinitionFlag string

const (
	FlagNoOverride LabelDefinitionFlag = "no-override"
	FlagAdult      LabelDefinitionFlag = "adult"
	FlagUnauthed   LabelDefinitionFlag = "unauthed"
	FlagNoSelf     LabelDefinitionFlag = "no-self"
)

type Behavior struct {
	ProfileList  BehaviorAction
	ProfileView  BehaviorAction
	Avatar       BehaviorAction
	Banner       BehaviorAction
	ContentList  BehaviorAction
	ContentView  BehaviorAction
	ContentMedia BehaviorAction
}

type LabelDefinition struct {
	Identifier     string
	DefaultSetting LabelPreference
	Configurable   bool
	Flags          []LabelDefinitionFlag
	Blurs          string // content, media, none
	Severity       string // alert, inform, none
	AdultOnly      bool
	Message        string
	Behaviors      map[LabelTarget]Behavior
}

func (d LabelDefinition) hasFlag(flag LabelDefinitionFlag) bool {
	for _, candidate := range d.Flags {
		if candidate == flag {
			return true
		}
	}
	return false
}

func alertOrInform(severity string) BehaviorAction {
	switch severity {
	case "alert":
		return BehaviorAlert
	case "inform":
		return BehaviorInform
	default:
		return ""
	}
}

func buildBehaviors(blurs, severity string, adultOnly bool) map[LabelTarget]Behavior {
	notice := alertOrInform(severity)
	account := Behavior{}
	profile := Behavior{}
	content := Behavior{}

	switch blurs {
	case "content":
		account.ProfileList = notice
		account.ProfileView = notice
		account.ContentList = BehaviorBlur
		if adultOnly {
			account.ContentView = BehaviorBlur
		} else {
			account.ContentView = notice
		}

		profile.ProfileList = notice
		profile.ProfileView = notice

		content.ContentList = BehaviorBlur
		if adultOnly {
			content.ContentView = BehaviorBlur
		} else {
			content.ContentView = notice
		}
	case "media":
		account.ProfileList = notice
		account.ProfileView = notice
		account.Avatar = BehaviorBlur
		account.Banner = BehaviorBlur

		profile.ProfileList = notice
		profile.ProfileView = notice
		profile.Avatar = BehaviorBlur
		profile.Banner = BehaviorBlur

		content.ContentMedia = BehaviorBlur
	case "none":
		account.ProfileList = notice
		account.ProfileView = notice
		account.ContentList = notice
		account.ContentView = notice

		profile.ProfileList = notice
		profile.ProfileView = notice

		content.ContentList = notice
		content.ContentView = notice
	}

	return map[LabelTarget]Behavior{
		LabelTargetAccount: account,
		LabelTargetProfile: profile,
		LabelTargetContent: content,
	}
}

func newLabelDefinition(
	identifier string,
	defaultSetting LabelPreference,
	configurable bool,
	flags []LabelDefinitionFlag,
	blurs, severity string,
	adultOnly bool,
	message string,
) LabelDefinition {
	return LabelDefinition{
		Identifier:     identifier,
		DefaultSetting: defaultSetting,
		Configurable:   configurable,
		Flags:          flags,
		Blurs:          blurs,
		Severity:       severity,
		AdultOnly:      adultOnly,
		Message:        message,
		Behaviors:      buildBehaviors(blurs, severity, adultOnly),
	}
}

var builtInLabels = map[string]LabelDefinition{
	"!hide": newLabelDefinition(
		"!hide", LabelHide, false,
		[]LabelDefinitionFlag{FlagNoOverride},
		"content", "alert", false,
		"Content blocked",
	),
	"!warn": newLabelDefinition(
		"!warn", LabelWarn, false,
		nil,
		"content", "alert", false,
		"Content warning",
	),
	"!no-unauthenticated": newLabelDefinition(
		"!no-unauthenticated", LabelHide, false,
		[]LabelDefinitionFlag{FlagUnauthed, FlagNoOverride},
		"content", "alert", false,
		"Sign in to view",
	),
	"porn": newLabelDefinition(
		"porn", LabelHide, true,
		[]LabelDefinitionFlag{FlagAdult},
		"media", "alert", true,
		"Adult content",
	),
	"sexual": newLabelDefinition(
		"sexual", LabelWarn, true,
		nil,
		"media", "alert", false,
		"Suggestive content",
	),
	"nudity": newLabelDefinition(
		"nudity", LabelIgnore, true,
		nil,
		"media", "inform", false,
		"Non-sexual nudity",
	),
	"graphic-media": newLabelDefinition(
		"graphic-media", LabelHide, true,
		[]LabelDefinitionFlag{FlagAdult},
		"media", "alert", true,
		"Graphic content",
	),
}

func lookupLabelDefinition(val string) (LabelDefinition, bool) {
	def, ok := builtInLabels[val]
	return def, ok
}

func (b Behavior) action(context UIContext) BehaviorAction {
	switch context {
	case UIContextContentList:
		return b.ContentList
	case UIContextContentView:
		return b.ContentView
	case UIContextContentMedia:
		return b.ContentMedia
	default:
		return ""
	}
}
