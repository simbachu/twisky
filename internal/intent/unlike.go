package intent

// UnlikePost deletes an app.bsky.feed.like record by AT URI.
type UnlikePost struct {
	Record string
}

func (UnlikePost) intent() {}
