package post

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"

	"github.com/rivo/uniseg"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/richtext"
)

const MaxPostGraphemes = 300

// MaxImageBytes is the AT Protocol limit for app.bsky.embed.images blobs.
const MaxImageBytes = 1_000_000

const MaxAltGraphemes = 2000

const (
	mentionFacetType = "app.bsky.richtext.facet#mention"
)

// Writer creates post records on the viewer's PDS and resolves handles for mentions.
type Writer interface {
	CreatePost(ctx context.Context, text string, facets []bluesky.Facet, reply *intent.ReplyTo, image *intent.PostImage) (string, error)
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
	image, err := prepareImage(i.Image)
	if err != nil {
		return "", err
	}
	if text == "" && image == nil {
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
	return h.writer.CreatePost(ctx, text, facets, i.Reply, image)
}

func prepareImage(img *intent.PostImage) (*intent.PostImage, error) {
	if img == nil || len(img.Data) == 0 {
		return nil, nil
	}
	if len(img.Data) > MaxImageBytes {
		return nil, fmt.Errorf("post: image exceeds %d bytes", MaxImageBytes)
	}
	alt := img.Alt
	if graphemeCount(alt) > MaxAltGraphemes {
		return nil, fmt.Errorf("post: alt exceeds %d graphemes", MaxAltGraphemes)
	}
	mime := http.DetectContentType(img.Data)
	switch mime {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
	default:
		return nil, fmt.Errorf("post: image type %q is not allowed", mime)
	}
	out := &intent.PostImage{
		Data: img.Data,
		MIME: mime,
		Alt:  alt,
	}
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(img.Data)); err == nil {
		out.Width = cfg.Width
		out.Height = cfg.Height
	}
	return out, nil
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
