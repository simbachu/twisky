package feed

import "github.com/simbachu/twisky/internal/bluesky"

// ApplyViewerLikes sets Liked and Reposted (and fills missing CID) on views from authenticated getPosts results.
func ApplyViewerLikes(views []PostView, posts []bluesky.Post) {
	if len(views) == 0 || len(posts) == 0 {
		return
	}
	byURI := make(map[string]bluesky.Post, len(posts))
	for _, post := range posts {
		if post.URI == "" {
			continue
		}
		byURI[post.URI] = post
	}
	for i := range views {
		applyViewerLike(&views[i], byURI)
	}
}

// ApplyViewerLikesToPage enriches the focused post, ancestors, and reply tree.
func ApplyViewerLikesToPage(view *PostPageView, posts []bluesky.Post) {
	if view == nil || len(posts) == 0 {
		return
	}
	byURI := make(map[string]bluesky.Post, len(posts))
	for _, post := range posts {
		if post.URI == "" {
			continue
		}
		byURI[post.URI] = post
	}
	applyViewerLike(&view.Post, byURI)
	for i := range view.Ancestors {
		applyViewerLike(&view.Ancestors[i].Post, byURI)
	}
	applyViewerLikesToReplies(view.Replies, byURI)
}

func applyViewerLikesToReplies(replies []ThreadNodeView, byURI map[string]bluesky.Post) {
	for i := range replies {
		applyViewerLike(&replies[i].Post, byURI)
		applyViewerLikesToReplies(replies[i].Replies, byURI)
	}
}

func applyViewerLike(view *PostView, byURI map[string]bluesky.Post) {
	if view == nil {
		return
	}
	post, ok := byURI[view.URI()]
	if !ok {
		return
	}
	view.Liked = post.Viewer != nil && post.Viewer.Like != ""
	view.Reposted = post.Viewer != nil && post.Viewer.Repost != ""
	if view.postCID == "" && post.CID != "" {
		view.postCID = post.CID
	}
	if view.QuotedPostMaybe != nil {
		applyViewerLike(view.QuotedPostMaybe, byURI)
	}
	if view.ReplyParentMaybe != nil {
		applyViewerLike(view.ReplyParentMaybe, byURI)
	}
}

// CollectPostURIs gathers AT-URIs from a feed for authenticated enrichment.
func CollectPostURIs(views []PostView) []string {
	seen := make(map[string]struct{})
	var uris []string
	var add func(v *PostView)
	add = func(v *PostView) {
		if v == nil {
			return
		}
		if uri := v.URI(); uri != "" {
			if _, ok := seen[uri]; !ok {
				seen[uri] = struct{}{}
				uris = append(uris, uri)
			}
		}
		if v.QuotedPostMaybe != nil {
			add(v.QuotedPostMaybe)
		}
		if v.ReplyParentMaybe != nil {
			add(v.ReplyParentMaybe)
		}
	}
	for i := range views {
		add(&views[i])
	}
	return uris
}

// CollectPagePostURIs gathers AT-URIs from a post page for authenticated enrichment.
func CollectPagePostURIs(view PostPageView) []string {
	seen := make(map[string]struct{})
	var uris []string
	var add func(v *PostView)
	add = func(v *PostView) {
		if v == nil {
			return
		}
		if uri := v.URI(); uri != "" {
			if _, ok := seen[uri]; !ok {
				seen[uri] = struct{}{}
				uris = append(uris, uri)
			}
		}
		if v.QuotedPostMaybe != nil {
			add(v.QuotedPostMaybe)
		}
		if v.ReplyParentMaybe != nil {
			add(v.ReplyParentMaybe)
		}
	}
	add(&view.Post)
	for i := range view.Ancestors {
		add(&view.Ancestors[i].Post)
	}
	var addReplies func(replies []ThreadNodeView)
	addReplies = func(replies []ThreadNodeView) {
		for i := range replies {
			add(&replies[i].Post)
			addReplies(replies[i].Replies)
		}
	}
	addReplies(view.Replies)
	return uris
}
