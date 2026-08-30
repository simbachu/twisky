package post

import (
	"context"
	"fmt"
	"strings"

	"github.com/rivo/uniseg"
	"github.com/simbachu/twisky/internal/intent"
)

const MaxPostGraphemes = 300

// Writer creates post records on the viewer's PDS.
type Writer interface {
	CreatePost(ctx context.Context, text string) (string, error)
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
	if graphemeCount(text) > MaxPostGraphemes {
		return "", fmt.Errorf("post: text exceeds %d graphemes", MaxPostGraphemes)
	}
	if h.writer == nil {
		return "", fmt.Errorf("post: writer is required")
	}
	return h.writer.CreatePost(ctx, text)
}

func graphemeCount(text string) int {
	return uniseg.GraphemeClusterCount(text)
}
