package repost

import (
	"context"
	"fmt"
	"strings"

	"github.com/simbachu/twisky/internal/intent"
)

// Writer creates repost records on the viewer's PDS.
type Writer interface {
	CreateRepost(ctx context.Context, uri, cid string) error
}

// Handler executes RepostPost intents.
type Handler struct {
	writer Writer
}

func NewHandler(writer Writer) *Handler {
	return &Handler{writer: writer}
}

func (h *Handler) Handle(ctx context.Context, i intent.RepostPost) error {
	uri := strings.TrimSpace(i.URI)
	cid := strings.TrimSpace(i.CID)
	if uri == "" || cid == "" {
		return fmt.Errorf("repost: uri and cid are required")
	}
	if h.writer == nil {
		return fmt.Errorf("repost: writer is required")
	}
	return h.writer.CreateRepost(ctx, uri, cid)
}
