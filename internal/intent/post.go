package intent

// ViewPost loads a single post and its thread for the given URL slug and record key.
type ViewPost struct {
	Slug string
	ID   string
}

func (ViewPost) intent() {}
