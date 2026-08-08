package intent

// ViewTag loads posts tagged with the given hashtag (URL slug under /tagged/).
type ViewTag struct {
	Tag       string
	Cursor    string
	// HeadCheck skips reply-parent and mention enrichment. Used for ?since=
	// polls that only need post IDs to detect new items.
	HeadCheck bool
}

func (ViewTag) intent() {}
