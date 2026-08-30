package intent

// ViewSettings loads the authenticated user's Bluesky settings.
type ViewSettings struct {
	DID    string
	Handle string
}

func (ViewSettings) intent() {}

// UpdateContentFiltering writes adult-content and global content-label preferences.
type UpdateContentFiltering struct {
	AdultContentEnabled bool
	Labels              map[string]string
}

func (UpdateContentFiltering) intent() {}

// UpdateThreading writes thread-view preferences.
type UpdateThreading struct {
	Sort                    string
	PrioritizeFollowedUsers bool
}

func (UpdateThreading) intent() {}
