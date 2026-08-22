package like

import (
	"context"
	"fmt"
	"strings"

	"github.com/simbachu/twisky/internal/intent"
)

// Writer creates and deletes like records on the viewer's PDS.
type Writer interface {
	CreateLike(ctx context.Context, uri, cid string) (string, error)
	DeleteLike(ctx context.Context, recordURI string) error
}

// Handler executes LikePost and UnlikePost intents.
type Handler struct {
	writer Writer
}

func NewHandler(writer Writer) *Handler {
	return &Handler{writer: writer}
}

func (h *Handler) HandleLike(ctx context.Context, i intent.LikePost) (string, error) {
	uri := strings.TrimSpace(i.URI)
	cid := strings.TrimSpace(i.CID)
	if uri == "" || cid == "" {
		return "", fmt.Errorf("like: uri and cid are required")
	}
	if h.writer == nil {
		return "", fmt.Errorf("like: writer is required")
	}
	return h.writer.CreateLike(ctx, uri, cid)
}

func (h *Handler) HandleUnlike(ctx context.Context, i intent.UnlikePost) error {
	record := strings.TrimSpace(i.Record)
	if record == "" {
		return fmt.Errorf("like: record is required")
	}
	if h.writer == nil {
		return fmt.Errorf("like: writer is required")
	}
	return h.writer.DeleteLike(ctx, record)
}
