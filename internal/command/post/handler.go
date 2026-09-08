package post

import (
	"context"
	"fmt"
	"strings"

	"github.com/rivo/uniseg"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/richtext"
)

const MaxPostGraphemes = 300

const (
	mentionFacetType = "app.bsky.richtext.facet#mention"
)

// Writer creates post records on the viewer's PDS and resolves handles for mentions.
type Writer interface {
	CreatePost(ctx context.Context, text string, facets []bluesky.Facet, reply *intent.ReplyTo) (string, error)
	ResolveHandle(ctx context.Context, handle string) (string, error)
}

// Handler executes CreatePost intents.
type Handler struct {
	writer Writer
}

func NewHandler(writer Writer) *Handler {
	return &Handler{writer: writer}
}

func (h *Handler) HandleCreate(ctx context.Context, i intent.CreatePost) (string, error) {
	text := strings.TrimSpace(i.Text)
	if text == "" {
		return "", fmt.Errorf("post: text is required")
	}
	if i.Reply != nil {
		if err := validateReply(i.Reply); err != nil {
			return "", err
		}
	}
	if h.writer == nil {
		return "", fmt.Errorf("post: writer is required")
	}

	facets := richtext.DetectFacets(text)
	text, facets = richtext.ShortenLinks(text, facets)
	facets = h.resolveMentions(ctx, facets)
	facets = richtext.StripInvalidMentions(facets)

	if graphemeCount(text) > MaxPostGraphemes {
		return "", fmt.Errorf("post: text exceeds %d graphemes", MaxPostGraphemes)
	}
	return h.writer.CreatePost(ctx, text, facets, i.Reply)
}

func (h *Handler) resolveMentions(ctx context.Context, facets []bluesky.Facet) []bluesky.Facet {
	if len(facets) == 0 {
		return facets
	}
	for i := range facets {
		for j := range facets[i].Features {
			feature := &facets[i].Features[j]
			if feature.Type != mentionFacetType {
				continue
			}
			if feature.DID == "" || strings.HasPrefix(feature.DID, "did:") {
				continue
			}
			did, err := h.writer.ResolveHandle(ctx, feature.DID)
			if err != nil || did == "" {
				feature.DID = ""
				continue
			}
			feature.DID = did
		}
	}
	return facets
}

func validateReply(reply *intent.ReplyTo) error {
	if reply.RootURI == "" || reply.RootCID == "" || reply.ParentURI == "" || reply.ParentCID == "" {
		return fmt.Errorf("post: reply refs are incomplete")
	}
	return nil
}

func graphemeCount(text string) int {
	return uniseg.GraphemeClusterCount(text)
}
