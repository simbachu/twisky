package command

import (
	"context"
	"fmt"

	"github.com/simbachu/twisky/internal/command/like"
	"github.com/simbachu/twisky/internal/intent"
)

// Dispatcher routes write intents to command handlers.
type Dispatcher struct{}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Dispatch routes i to the matching command handler.
// writer is the per-request authenticated PDS client (required for LikePost).
func (d *Dispatcher) Dispatch(ctx context.Context, i intent.Intent, writer like.Writer) error {
	switch i := i.(type) {
	case intent.LikePost:
		return like.NewHandler(writer).Handle(ctx, i)
	default:
		return fmt.Errorf("command: unknown intent %T", i)
	}
}
