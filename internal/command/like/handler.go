package like

import (
	"context"
	"fmt"
	"strings"

	"github.com/simbachu/twisky/internal/intent"
)

// Writer creates like records on the viewer's PDS.
type Writer interface {
	CreateLike(ctx context.Context, uri, cid string) error
}

// Handler executes LikePost intents.
type Handler struct {
	writer Writer
}

func NewHandler(writer Writer) *Handler {
	return &Handler{writer: writer}
}

func (h *Handler) Handle(ctx context.Context, i intent.LikePost) error {
	uri := strings.TrimSpace(i.URI)
	cid := strings.TrimSpace(i.CID)
	if uri == "" || cid == "" {
		return fmt.Errorf("like: uri and cid are required")
	}
	if h.writer == nil {
		return fmt.Errorf("like: writer is required")
	}
	return h.writer.CreateLike(ctx, uri, cid)
}
