package intent

// CreatePost writes a new app.bsky.feed.post record.
// Facets (links, tags, mentions) are detected server-side on create.
type CreatePost struct {
	Text  string
	Reply *ReplyTo
	// Image, when set, is uploaded and attached as app.bsky.embed.images.
	Image *PostImage
}

// PostImage is one image to attach to a new post.
type PostImage struct {
	Data   []byte
	MIME   string
	Alt    string
	Width  int // 0 if unknown
	Height int
}

// ReplyTo holds AT Protocol reply root and parent strong refs.
type ReplyTo struct {
	RootURI   string
	RootCID   string
	ParentURI string
	ParentCID string
}

func (CreatePost) intent() {}
