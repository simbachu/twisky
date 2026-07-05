package intent

// ProfileTab identifies which section of a profile page to show.
type ProfileTab string

const (
	ProfileTabPosts ProfileTab = "posts"
	ProfileTabMedia ProfileTab = "media"
)

// ViewProfile loads a Bluesky actor profile for the given URL slug.
// Slug must be a valid handle (e.g. bsky.app) or DID (e.g. did:plc:...).
type ViewProfile struct {
	Slug string
	Tab  ProfileTab
}

func (ViewProfile) intent() {}
