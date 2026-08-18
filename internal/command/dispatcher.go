package command

import (
	"context"
	"fmt"

	"github.com/simbachu/twisky/internal/command/like"
	"github.com/simbachu/twisky/internal/command/repost"
	"github.com/simbachu/twisky/internal/intent"
)

// Dispatcher routes write intents to command handlers.
type Dispatcher struct{}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Dispatch routes i to the matching command handler.
// writer is the per-request authenticated PDS client: a like.Writer for
// LikePost, a repost.Writer for RepostPost.
func (d *Dispatcher) Dispatch(ctx context.Context, i intent.Intent, writer any) error {
	switch i := i.(type) {
	case intent.LikePost:
		w, ok := writer.(like.Writer)
		if !ok {
			return fmt.Errorf("command: like writer is required")
		}
		return like.NewHandler(w).Handle(ctx, i)
	case intent.RepostPost:
		w, ok := writer.(repost.Writer)
		if !ok {
			return fmt.Errorf("command: repost writer is required")
		}
		return repost.NewHandler(w).Handle(ctx, i)
	default:
		return fmt.Errorf("command: unknown intent %T", i)
	}
}
