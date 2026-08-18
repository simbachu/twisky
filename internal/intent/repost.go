package intent

// RepostPost creates an app.bsky.feed.repost record for the given post strong-ref.
type RepostPost struct {
	URI string
	CID string
}

func (RepostPost) intent() {}
