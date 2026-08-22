package repost

import (
	"context"
	"fmt"
	"strings"

	"github.com/simbachu/twisky/internal/intent"
)

// Writer creates and deletes repost records on the viewer's PDS.
type Writer interface {
	CreateRepost(ctx context.Context, uri, cid string) (string, error)
	DeleteRepost(ctx context.Context, recordURI string) error
}

// Handler executes RepostPost and UnrepostPost intents.
type Handler struct {
	writer Writer
}

func NewHandler(writer Writer) *Handler {
	return &Handler{writer: writer}
}

func (h *Handler) HandleRepost(ctx context.Context, i intent.RepostPost) (string, error) {
	uri := strings.TrimSpace(i.URI)
	cid := strings.TrimSpace(i.CID)
	if uri == "" || cid == "" {
		return "", fmt.Errorf("repost: uri and cid are required")
	}
	if h.writer == nil {
		return "", fmt.Errorf("repost: writer is required")
	}
	return h.writer.CreateRepost(ctx, uri, cid)
}

func (h *Handler) HandleUnrepost(ctx context.Context, i intent.UnrepostPost) error {
	record := strings.TrimSpace(i.Record)
	if record == "" {
		return fmt.Errorf("repost: record is required")
	}
	if h.writer == nil {
		return fmt.Errorf("repost: writer is required")
	}
	return h.writer.DeleteRepost(ctx, record)
}
