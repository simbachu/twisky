package intent

// ViewPost loads a single post and its thread for the given URL slug and record key.
// Part is "" for the full page or feedquery.PostPagePartAncestors for the ancestors fragment.
type ViewPost struct {
	Slug string
	ID   string
	Part string
}

func (ViewPost) intent() {}
