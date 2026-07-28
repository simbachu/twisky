package actor

import (
	"strings"
)

const (
	AvatarEmojiShield    = "🛡"
	AvatarEmojiButterfly = "🦋"
	BlueskyHandle        = "bsky.app"
)

// match handle to bsky.app domain, regardless of subdomain
func IsBlueskyHandle(handle string) bool {
	return strings.HasSuffix(strings.ToLower(handle), "."+BlueskyHandle)
}

// IsLabelerAccount reports whether an actor is a moderation labeler service.
func IsLabelerAccount(handle, did string, labelerFromAPI bool) bool {
	if labelerFromAPI {
		return true
	}
	return false
}

// AvatarFallbackEmoji returns a stand-in avatar glyph when no image URL exists.
// Bluesky handles are identified by the bsky.app domain, regardless of subdomain.
// Labeler accounts are identified by the moderation.bsky.app domain.
// Apply Bluesky first, then labeler account.
func AvatarFallbackEmoji(handle, did string, isLabeler bool) string {
	if IsBlueskyHandle(handle) {
		return AvatarEmojiButterfly
	}
	if IsLabelerAccount(handle, did, isLabeler) {
		return AvatarEmojiShield
	}
	return ""
}
