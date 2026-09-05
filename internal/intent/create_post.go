package intent

// CreatePost writes a new app.bsky.feed.post record with plain text.
type CreatePost struct {
	Text  string
	Reply *ReplyTo
}

// ReplyTo holds AT Protocol reply root and parent strong refs.
type ReplyTo struct {
	RootURI   string
	RootCID   string
	ParentURI string
	ParentCID string
}

func (CreatePost) intent() {}
