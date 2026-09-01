package intent

// ViewThread loads an author's contiguous thought-thread from the root post.
type ViewThread struct {
	Slug string
	ID   string
}

func (ViewThread) intent() {}
