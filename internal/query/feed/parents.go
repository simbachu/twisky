package feed

import (
	"context"

	"github.com/simbachu/twisky/internal/bluesky"
)

type PostFetcher interface {
	GetPosts(ctx context.Context, uris []string) ([]bluesky.Post, error)
}

// EnrichReplyParents hydrates reply parents for feed posts that have a reply ref but no parent view.
func EnrichReplyParents(ctx context.Context, fetcher PostFetcher, view FeedView) (FeedView, error) {
	if fetcher == nil {
		return view, nil
	}

	uris := collectReplyParentURIs(view.Posts)
	if len(uris) == 0 {
		return view, nil
	}

	posts, err := fetcher.GetPosts(ctx, uris)
	if err != nil {
		return FeedView{}, err
	}

	parentByURI := make(map[string]PostView, len(posts))
	for _, post := range posts {
		parentByURI[post.URI] = InsetPostView(NewPostView(post))
	}

	enriched := make([]PostView, len(view.Posts))
	for i, post := range view.Posts {
		enriched[i] = post
		if post.ReplyParentMaybe != nil {
			continue
		}
		parent, ok := parentByURI[post.replyParentURI]
		if !ok {
			continue
		}
		enriched[i].ReplyParentMaybe = &parent
	}

	return FeedView{
		Posts:      enriched,
		NextCursor: view.NextCursor,
	}, nil
}

func collectReplyParentURIs(posts []PostView) []string {
	seen := make(map[string]struct{})
	uris := make([]string, 0)
	for _, post := range posts {
		if post.ReplyParentMaybe != nil || post.replyParentURI == "" {
			continue
		}
		if _, ok := seen[post.replyParentURI]; ok {
			continue
		}
		seen[post.replyParentURI] = struct{}{}
		uris = append(uris, post.replyParentURI)
	}
	return uris
}

// InsetPostView returns a single-level post view suitable for inset cards in feeds.
func InsetPostView(view PostView) PostView {
	view.ReplyParentMaybe = nil
	view.QuotedPostMaybe = nil
	view.replyParentURI = ""
	return view
}
