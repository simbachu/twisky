package oauth

// DefaultScopes are requested on every OAuth login and embedded in the public client ID.
// Users must log out and back in after these change.
//
// authViewAll covers timeline, saved-feed preferences, posts, and profiles via AppView proxy.
// Separate repo create/delete scopes cover likes and reposts (repo actions must be
// create/update/delete — "manage" is invalid and causes invalid_scope on PAR).
var DefaultScopes = []string{
	"atproto",
	"include:app.bsky.authViewAll?aud=did:web:api.bsky.app%23bsky_appview",
	"rpc:app.bsky.actor.putPreferences?aud=*",
	"repo:app.bsky.feed.like?action=create",
	"repo:app.bsky.feed.like?action=delete",
	"repo:app.bsky.feed.repost?action=create",
	"repo:app.bsky.feed.repost?action=delete",
	"repo:app.bsky.feed.post?action=create",
}
