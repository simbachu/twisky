package feed

import (
	"sort"

	"github.com/simbachu/twisky/internal/bluesky"
)

type ThreadNodeView struct {
	Post        PostView
	Unavailable bool
	Replies     []ThreadNodeView
}

type AncestorNodeView struct {
	Post        PostView
	Unavailable bool
}

const (
	PostPagePartAncestors = "ancestors"
	PostPagePartCounts    = "counts"
)

type PostPageView struct {
	Post             PostView
	Ancestors        []AncestorNodeView
	Replies          []ThreadNodeView
	HasAncestors     bool
	ReplyParentMaybe *AuthorView
	// ExplicitLive is set by the HTTP layer from the ?live=1 query param on
	// full page loads, so a shared link can restart live counts polling even
	// for a post that wouldn't otherwise auto-start it.
	ExplicitLive bool
}

func (PostPageView) IsResponse() {}

func NewPostPageView(root bluesky.ThreadViewPost, part string) PostPageView {
	if part == PostPagePartAncestors {
		return PostPageView{
			Ancestors: CollectAncestorNodes(root),
		}
	}
	ancestors := CollectAncestorNodes(root)
	return PostPageView{
		Post:             NewPostView(root.Post),
		HasAncestors:     len(ancestors) > 0,
		ReplyParentMaybe: replyParentAuthor(ancestors),
		Replies:          NewThreadNodeViews(root.Replies),
	}
}

// replyParentAuthor returns the immediate parent author from furthest-first ancestors.
func replyParentAuthor(ancestors []AncestorNodeView) *AuthorView {
	if len(ancestors) == 0 {
		return nil
	}
	immediate := ancestors[len(ancestors)-1]
	if immediate.Unavailable {
		return nil
	}
	author := AuthorView{
		Handle:      immediate.Post.AuthorHandle,
		DisplayName: immediate.Post.AuthorDisplayName,
		Avatar:      immediate.Post.AuthorAvatar,
	}
	return &author
}

func CollectAncestorNodes(root bluesky.ThreadViewPost) []AncestorNodeView {
	ancestors := make([]AncestorNodeView, 0)
	for node := root.Parent; node != nil; {
		switch typed := node.(type) {
		case bluesky.ThreadViewPost:
			ancestors = append(ancestors, AncestorNodeView{
				Post: NewPostView(typed.Post),
			})
			node = typed.Parent
		case bluesky.NotFoundPost, bluesky.BlockedPost:
			ancestors = append(ancestors, AncestorNodeView{Unavailable: true})
			return reverseAncestorNodes(ancestors)
		default:
			return reverseAncestorNodes(ancestors)
		}
	}
	return reverseAncestorNodes(ancestors)
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
	sortThreadNodesOldestFirst(&views)
	return views
}

func sortThreadNodesOldestFirst(nodes *[]ThreadNodeView) {
	available := make([]ThreadNodeView, 0, len(*nodes))
	unavailable := make([]ThreadNodeView, 0)
	for _, node := range *nodes {
		if node.Unavailable {
			unavailable = append(unavailable, node)
			continue
		}
		sortThreadNodesOldestFirst(&node.Replies)
		available = append(available, node)
	}
	sort.Slice(available, func(i, j int) bool {
		return available[i].Post.CreatedAt.Before(available[j].Post.CreatedAt)
	})
	*nodes = append(available, unavailable...)
}

func reversePostViews(views []PostView) {
	for i, j := 0, len(views)-1; i < j; i, j = i+1, j-1 {
		views[i], views[j] = views[j], views[i]
	}
}

func reverseAncestorNodes(nodes []AncestorNodeView) []AncestorNodeView {
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
	return nodes
}

func collectPostsFromPage(view PostPageView) []PostView {
	posts := make([]PostView, 0, 1+len(view.Ancestors))
	for _, ancestor := range view.Ancestors {
		if ancestor.Unavailable {
			continue
		}
		posts = append(posts, ancestor.Post)
	}
	if view.Post.ID != "" {
		posts = append(posts, view.Post)
	}
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
