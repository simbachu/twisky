package intent

// CreatePost writes a new app.bsky.feed.post record with plain text.
type CreatePost struct {
	Text string
}

func (CreatePost) intent() {}
