package post

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PostSpine wraps connected posts in ul.post-spine. Each item is wrapped in li.
// id is optional (e.g. thought-thread-list).
func PostSpine(id string, items []g.Node, extra ...g.Node) g.Node {
	attrs := []g.Node{g.Attr("class", "post-spine")}
	if id != "" {
		attrs = append(attrs, g.Attr("id", id))
	}
	attrs = append(attrs, extra...)

	lis := make([]g.Node, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		lis = append(lis, Li(item))
	}
	return Ul(g.Group(attrs), g.Group(lis))
}
