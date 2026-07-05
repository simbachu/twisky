package intent

// ViewTag loads posts tagged with the given hashtag (URL slug under /tagged/).
type ViewTag struct {
	Tag string
}

func (ViewTag) intent() {}
