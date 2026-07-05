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
		for _, segment := range post.TextSegments {
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
		posts[i] = post
		if len(post.TextSegments) == 0 {
			continue
		}
		segments := make([]richtext.Segment, len(post.TextSegments))
		copy(segments, post.TextSegments)
		for j, segment := range segments {
			if segment.Kind != richtext.Mention {
				continue
			}
			if handle, ok := handleByDID[segment.Mention]; ok {
				segments[j].Mention = handle
			}
		}
		posts[i].TextSegments = segments
	}
	return FeedView{
		Posts:      posts,
		NextCursor: view.NextCursor,
	}
}
