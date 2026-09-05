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
	CreatePost(ctx context.Context, text string, reply *intent.ReplyTo) (string, error)
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
	if i.Reply != nil {
		if err := validateReply(i.Reply); err != nil {
			return "", err
		}
	}
	if h.writer == nil {
		return "", fmt.Errorf("post: writer is required")
	}
	return h.writer.CreatePost(ctx, text, i.Reply)
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
