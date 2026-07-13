package feed

import "context"

func ResolveMentionHandlesInThread(ctx context.Context, resolver ProfileResolver, view PostPageView) PostPageView {
	if resolver == nil {
		return view
	}

	posts := collectPostsFromPage(view)
	feedView := FeedView{Posts: posts}
	resolved := ResolveMentionHandles(ctx, resolver, feedView)
	postByID := make(map[string]PostView, len(resolved.Posts))
	for _, post := range resolved.Posts {
		postByID[post.ID] = post
	}

	view.Ancestors = rewriteAncestorNodes(view.Ancestors, postByID)
	view.Post = postByID[view.Post.ID]
	view.Replies = rewriteThreadNodes(view.Replies, postByID)
	return view
}

func rewriteAncestorNodes(nodes []AncestorNodeView, postByID map[string]PostView) []AncestorNodeView {
	rewritten := make([]AncestorNodeView, len(nodes))
	for i, node := range nodes {
		rewritten[i] = node
		if node.Unavailable {
			continue
		}
		if resolved, ok := postByID[node.Post.ID]; ok {
			rewritten[i].Post = resolved
		}
	}
	return rewritten
}

func rewritePostViews(posts []PostView, postByID map[string]PostView) []PostView {
	rewritten := make([]PostView, len(posts))
	for i, post := range posts {
		if resolved, ok := postByID[post.ID]; ok {
			rewritten[i] = resolved
		} else {
			rewritten[i] = post
		}
	}
	return rewritten
}

func rewriteThreadNodes(nodes []ThreadNodeView, postByID map[string]PostView) []ThreadNodeView {
	rewritten := make([]ThreadNodeView, len(nodes))
	for i, node := range nodes {
		rewritten[i] = node
		if node.Unavailable {
			continue
		}
		if resolved, ok := postByID[node.Post.ID]; ok {
			rewritten[i].Post = resolved
		}
		rewritten[i].Replies = rewriteThreadNodes(node.Replies, postByID)
	}
	return rewritten
}
