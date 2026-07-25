package moderation

import "strings"

const profileRecordSuffix = "/app.bsky.actor.profile/self"

func IsProfileLabel(label Label) bool {
	return strings.HasSuffix(label.URI, profileRecordSuffix)
}

func IsAccountAuthorLabel(label Label) bool {
	return !IsProfileLabel(label) || label.Val == "!no-unauthenticated"
}

func LabelAppliesToPost(label Label, postURI, authorDID string) bool {
	if label.URI == "" {
		return true
	}
	if label.URI == postURI {
		return true
	}
	if authorDID != "" && label.URI == authorDID {
		return true
	}
	return false
}
