package intent

// ViewHome loads the authenticated user's following timeline.
type ViewHome struct {
	Cursor string
}

func (ViewHome) intent() {}
