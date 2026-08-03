package intent

// LikePost creates an app.bsky.feed.like record for the given post strong-ref.
type LikePost struct {
	URI string
	CID string
}

func (LikePost) intent() {}
