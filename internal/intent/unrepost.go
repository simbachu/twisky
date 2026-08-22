package intent

// UnrepostPost deletes an app.bsky.feed.repost record by AT URI.
type UnrepostPost struct {
	Record string
}

func (UnrepostPost) intent() {}
