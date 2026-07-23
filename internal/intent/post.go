package intent

// ViewPost loads a single post and its thread for the given URL slug and record key.
// Part is "" for the full page, or a feedquery.PostPagePart* constant for a
// fragment (ancestors, counts, or replies).
type ViewPost struct {
	Slug string
	ID   string
	Part string
}

func (ViewPost) intent() {}
