package feed

import (
	"github.com/simbachu/twisky/internal/atproto"
	"github.com/simbachu/twisky/internal/bluesky"
)

// ValidOPThreadNumber reports whether index and count describe a contiguous
// OP thought-thread position from AppView feed numbering.
func ValidOPThreadNumber(index, count int) bool {
	return index >= 1 && count >= 1 && index <= count
}

// ThoughtThreadView is the read model for an author's contiguous thought-thread.
type ThoughtThreadView struct {
	Posts      []PostView
	RootHandle string
	RootDID    string
	RootID     string
}

func (ThoughtThreadView) IsResponse() {}

// CollectOPThoughtSpine walks same-author replies from root and returns the
// contiguous OP chain (root first).
func CollectOPThoughtSpine(root bluesky.ThreadViewPost) []PostView {
	rootDID := root.Post.Author.DID
	spine := []PostView{NewPostView(root.Post)}

	node := root
	for {
		var next *bluesky.ThreadViewPost
		for _, reply := range node.Replies {
			typed, ok := reply.(bluesky.ThreadViewPost)
			if !ok || typed.Post.Author.DID != rootDID {
				continue
			}
			if next == nil || typed.Post.Record.CreatedAt.Before(next.Post.Record.CreatedAt) {
				candidate := typed
				next = &candidate
			}
		}
		if next == nil {
			break
		}
		spine = append(spine, NewPostView(next.Post))
		node = *next
	}
	return spine
}

func applyThreadRoot(view *PostView, rootURI string, rootPost *bluesky.Post) {
	if rootURI == "" {
		rootURI = view.postURI
	}
	id, err := atproto.PostRkey(rootURI)
	if err != nil {
		return
	}
	did, err := atproto.PostAuthorDID(rootURI)
	if err != nil {
		return
	}
	view.threadRootID = id
	view.threadRootDID = did
	if rootPost != nil {
		view.threadRootHandle = rootPost.Author.Handle
	}
}

func applyOPThreadNumbering(view *PostView, index, count int) {
	if !ValidOPThreadNumber(index, count) {
		return
	}
	view.opThreadIndex = index
	view.opThreadCount = count
}

// OPThreadNumber returns the 1-based position and total when AppView numbering
// is present and valid on this post.
func (v PostView) OPThreadNumber() (index, count int, ok bool) {
	if !ValidOPThreadNumber(v.opThreadIndex, v.opThreadCount) {
		return 0, 0, false
	}
	return v.opThreadIndex, v.opThreadCount, true
}

// ThreadRootPathParts returns handle, DID, and rkey for the conversation root.
func (v PostView) ThreadRootPathParts() (handle, did, rootID string) {
	handle = v.threadRootHandle
	did = v.threadRootDID
	rootID = v.threadRootID
	if rootID == "" {
		rootID = v.ID
	}
	if did == "" {
		did = v.authorDID
	}
	if handle == "" {
		handle = v.AuthorHandle
	}
	return handle, did, rootID
}

// PostViewWithOPThreadNumber sets OP thread numbering for tests outside this package.
func PostViewWithOPThreadNumber(view PostView, index, count int) PostView {
	applyOPThreadNumbering(&view, index, count)
	return view
}

// PostViewWithThreadRoot sets thread root path parts for tests outside this package.
func PostViewWithThreadRoot(view PostView, handle, did, rootID string) PostView {
	view.threadRootHandle = handle
	view.threadRootDID = did
	view.threadRootID = rootID
	return view
}
