package intent

// ViewHome loads the authenticated user's following timeline.
type ViewHome struct {
	Cursor string
	// HeadCheck skips reply-parent and mention enrichment. Used for ?since=
	// polls that only need post IDs to detect new items.
	HeadCheck bool
}

func (ViewHome) intent() {}
