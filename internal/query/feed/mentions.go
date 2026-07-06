package feed

import (
	"context"
	"strings"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/richtext"
)

const (
	maxResolvedMentions = 100
	profilesBatchSize = 25
)

type ProfileResolver interface {
	GetProfiles(ctx context.Context, actors []string) ([]bluesky.Profile, error)
}

func ResolveMentionHandles(ctx context.Context, resolver ProfileResolver, view FeedView) FeedView {
	if resolver == nil {
		return view
	}

	dids := collectMentionDIDs(view)
	if len(dids) == 0 {
		return view
	}

	handleByDID := resolveDIDsToHandles(ctx, resolver, dids)
	if len(handleByDID) == 0 {
		return view
	}

	return rewriteMentionHandles(view, handleByDID)
}

func collectMentionDIDs(view FeedView) []string {
	seen := make(map[string]struct{})
	dids := make([]string, 0)
	for _, post := range view.Posts {
		dids = appendMentionDIDsFromPost(post, seen, dids)
		if len(dids) >= maxResolvedMentions {
			return dids[:maxResolvedMentions]
		}
	}
	return dids
}

func appendMentionDIDsFromPost(post PostView, seen map[string]struct{}, dids []string) []string {
	dids = appendMentionDIDsFromSegments(post.TextSegments, seen, dids)
	if post.QuotedPostMaybe != nil {
		dids = appendMentionDIDsFromSegments(post.QuotedPostMaybe.TextSegments, seen, dids)
	}
	if post.ReplyParentMaybe != nil {
		dids = appendMentionDIDsFromSegments(post.ReplyParentMaybe.TextSegments, seen, dids)
	}
	return dids
}

func appendMentionDIDsFromSegments(segments []richtext.Segment, seen map[string]struct{}, dids []string) []string {
	for _, segment := range segments {
		if segment.Kind != richtext.Mention {
			continue
		}
		if !strings.HasPrefix(segment.Mention, "did:") {
			continue
		}
		if _, ok := seen[segment.Mention]; ok {
			continue
		}
		seen[segment.Mention] = struct{}{}
		dids = append(dids, segment.Mention)
		if len(dids) >= maxResolvedMentions {
			return dids
		}
	}
	return dids
}

func resolveDIDsToHandles(ctx context.Context, resolver ProfileResolver, dids []string) map[string]string {
	handleByDID := make(map[string]string)
	for start := 0; start < len(dids); start += profilesBatchSize {
		end := start + profilesBatchSize
		if end > len(dids) {
			end = len(dids)
		}
		chunk := dids[start:end]

		profiles, err := resolver.GetProfiles(ctx, chunk)
		if err != nil {
			continue
		}
		for _, profile := range profiles {
			if profile.DID == "" || profile.Handle == "" {
				continue
			}
			handleByDID[profile.DID] = profile.Handle
		}
	}
	return handleByDID
}

func rewriteMentionHandles(view FeedView, handleByDID map[string]string) FeedView {
	posts := make([]PostView, len(view.Posts))
	for i, post := range view.Posts {
		posts[i] = rewritePostMentionHandles(post, handleByDID)
	}
	return FeedView{
		Posts:      posts,
		NextCursor: view.NextCursor,
	}
}

func rewritePostMentionHandles(post PostView, handleByDID map[string]string) PostView {
	post.TextSegments = rewriteSegments(post.TextSegments, handleByDID)
	if post.QuotedPostMaybe != nil {
		quoted := rewritePostMentionHandles(*post.QuotedPostMaybe, handleByDID)
		post.QuotedPostMaybe = &quoted
	}
	if post.ReplyParentMaybe != nil {
		parent := rewritePostMentionHandles(*post.ReplyParentMaybe, handleByDID)
		post.ReplyParentMaybe = &parent
	}
	return post
}

func rewriteSegments(segments []richtext.Segment, handleByDID map[string]string) []richtext.Segment {
	if len(segments) == 0 {
		return segments
	}
	rewritten := make([]richtext.Segment, len(segments))
	copy(rewritten, segments)
	for i, segment := range rewritten {
		if segment.Kind != richtext.Mention {
			continue
		}
		if handle, ok := handleByDID[segment.Mention]; ok {
			rewritten[i].Mention = handle
		}
	}
	return rewritten
}
