package oauth

// DefaultScopes are requested on every OAuth login and embedded in the public client ID.
// Users must log out and back in after these change.
//
// transition:generic no longer grants AppView proxy access to getTimeline on many PDSes.
// authViewAll covers timeline, custom feeds, posts, and profiles; authManagePrefs covers
// saved-feed tabs; repo manage scopes cover likes and reposts.
var DefaultScopes = []string{
	"atproto",
	"include:app.bsky.authViewAll?aud=did:web:api.bsky.app%23bsky_appview",
	"include:app.bsky.authManagePrefs",
	"repo:app.bsky.feed.like?action=manage",
	"repo:app.bsky.feed.repost?action=manage",
}
