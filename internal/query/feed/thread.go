package feed

import (
	"github.com/simbachu/twisky/internal/bluesky"
)

type ThreadNodeView struct {
	Post         PostView
	Unavailable  bool
	Replies      []ThreadNodeView
}

type PostPageView struct {
	Post      PostView
	Ancestors []PostView
	Replies   []ThreadNodeView
}

func (PostPageView) IsResponse() {}

func NewPostPageView(root bluesky.ThreadViewPost) PostPageView {
	return PostPageView{
		Post:      NewPostView(root.Post),
		Ancestors: CollectAncestors(root),
		Replies:   NewThreadNodeViews(root.Replies),
	}
}

func CollectAncestors(root bluesky.ThreadViewPost) []PostView {
	ancestors := make([]PostView, 0)
	for node := root.Parent; node != nil; {
		thread, ok := node.(bluesky.ThreadViewPost)
		if !ok {
			break
		}
		ancestors = append(ancestors, NewPostView(thread.Post))
		node = thread.Parent
	}
	reversePostViews(ancestors)
	return ancestors
}

func NewThreadNodeViews(nodes []bluesky.ThreadNode) []ThreadNodeView {
	views := make([]ThreadNodeView, 0, len(nodes))
	for _, node := range nodes {
		switch typed := node.(type) {
		case bluesky.ThreadViewPost:
			views = append(views, ThreadNodeView{
				Post:    NewPostView(typed.Post),
				Replies: NewThreadNodeViews(typed.Replies),
			})
		case bluesky.NotFoundPost, bluesky.BlockedPost:
			views = append(views, ThreadNodeView{Unavailable: true})
		}
	}
	return views
}

func reversePostViews(views []PostView) {
	for i, j := 0, len(views)-1; i < j; i, j = i+1, j-1 {
		views[i], views[j] = views[j], views[i]
	}
}

func collectPostsFromPage(view PostPageView) []PostView {
	posts := make([]PostView, 0, 1+len(view.Ancestors))
	posts = append(posts, view.Ancestors...)
	posts = append(posts, view.Post)
	posts = append(posts, collectPostsFromThreadNodes(view.Replies)...)
	return posts
}

func collectPostsFromThreadNodes(nodes []ThreadNodeView) []PostView {
	posts := make([]PostView, 0)
	for _, node := range nodes {
		if node.Unavailable {
			continue
		}
		posts = append(posts, node.Post)
		posts = append(posts, collectPostsFromThreadNodes(node.Replies)...)
	}
	return posts
}
