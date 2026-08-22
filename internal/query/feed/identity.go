package feed

import (
	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/identity"
)

// ObserveAuthors records handle↔DID mappings from feed items.
func ObserveAuthors(dir *identity.Directory, view FeedView) {
	if dir == nil {
		return
	}
	for _, post := range view.Posts {
		observePostView(dir, post)
	}
}

func observePostView(dir *identity.Directory, view PostView) {
	dir.Observe(view.AuthorHandle, view.AuthorDID())
	if view.RepostedByMaybe != nil {
		// RepostedBy AuthorView has no DID; handle-only observation skipped.
	}
	if view.ReplyParentMaybe != nil {
		observePostView(dir, *view.ReplyParentMaybe)
	}
	if view.QuotedPostMaybe != nil {
		observePostView(dir, *view.QuotedPostMaybe)
	}
}

// ObservePost records mapping from a bluesky post payload.
func ObservePost(dir *identity.Directory, post bluesky.Post) {
	if dir == nil {
		return
	}
	dir.ObserveAuthor(post.Author)
}

// ApplyIdentity rewrites author handles using last-known-good mappings.
func ApplyIdentity(dir *identity.Directory, view PostView) PostView {
	if dir == nil {
		return view
	}
	ObservePost(dir, bluesky.Post{
		URI: "",
		Author: bluesky.Author{
			DID:         view.AuthorDID(),
			Handle:      view.AuthorHandle,
			DisplayName: view.AuthorDisplayName,
			Avatar:      view.AuthorAvatar,
		},
	})
	view.AuthorHandle = dir.DisplayHandle(view.AuthorHandle, view.AuthorDID())
	view.AuthorDisplayName = actor.Name(view.AuthorDisplayName, view.AuthorHandle)
	if view.RepostedByMaybe != nil {
		reposted := *view.RepostedByMaybe
		reposted.DisplayName = actor.Name(reposted.DisplayName, reposted.Handle)
		view.RepostedByMaybe = &reposted
	}
	if view.ReplyParentMaybe != nil {
		parent := ApplyIdentity(dir, *view.ReplyParentMaybe)
		view.ReplyParentMaybe = &parent
	}
	if view.QuotedPostMaybe != nil {
		quoted := ApplyIdentity(dir, *view.QuotedPostMaybe)
		view.QuotedPostMaybe = &quoted
	}
	return view
}

// ApplyIdentityFeed rewrites handles for every post in a feed view.
func ApplyIdentityFeed(dir *identity.Directory, view FeedView) FeedView {
	if dir == nil {
		return view
	}
	posts := make([]PostView, len(view.Posts))
	for i, post := range view.Posts {
		posts[i] = ApplyIdentity(dir, post)
	}
	return FeedView{
		Posts:      posts,
		NextCursor: view.NextCursor,
	}
}

// ApplyIdentityToPostPage rewrites handles across a post page view.
func ApplyIdentityToPostPage(dir *identity.Directory, view PostPageView) PostPageView {
	if dir == nil {
		return view
	}
	view.Post = ApplyIdentity(dir, view.Post)
	ancestors := make([]AncestorNodeView, len(view.Ancestors))
	for i, ancestor := range view.Ancestors {
		ancestors[i] = ancestor
		if !ancestor.Unavailable {
			ancestors[i].Post = ApplyIdentity(dir, ancestor.Post)
		}
	}
	view.Ancestors = ancestors
	replies := make([]ThreadNodeView, len(view.Replies))
	for i, reply := range view.Replies {
		replies[i] = applyIdentityThreadNode(dir, reply)
	}
	view.Replies = replies
	if view.ReplyParentMaybe != nil {
		parent := *view.ReplyParentMaybe
		parent.Handle = dir.DisplayHandle(parent.Handle, "")
		view.ReplyParentMaybe = &parent
	}
	return view
}

func applyIdentityThreadNode(dir *identity.Directory, node ThreadNodeView) ThreadNodeView {
	node.Post = ApplyIdentity(dir, node.Post)
	replies := make([]ThreadNodeView, len(node.Replies))
	for i, reply := range node.Replies {
		replies[i] = applyIdentityThreadNode(dir, reply)
	}
	node.Replies = replies
	return node
}
