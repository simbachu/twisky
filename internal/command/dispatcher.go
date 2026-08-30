package command

import (
	"context"
	"fmt"

	"github.com/simbachu/twisky/internal/command/like"
	"github.com/simbachu/twisky/internal/command/post"
	"github.com/simbachu/twisky/internal/command/repost"
	"github.com/simbachu/twisky/internal/command/settings"
	"github.com/simbachu/twisky/internal/intent"
)

// Dispatcher routes write intents to command handlers.
type Dispatcher struct{}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Dispatch routes i to the matching command handler.
// writer is the per-request authenticated PDS client.
// For create intents the returned string is the new record AT URI; for deletes it is empty.
func (d *Dispatcher) Dispatch(ctx context.Context, i intent.Intent, writer any) (string, error) {
	switch i := i.(type) {
	case intent.LikePost:
		w, ok := writer.(like.Writer)
		if !ok {
			return "", fmt.Errorf("command: like writer is required")
		}
		return like.NewHandler(w).HandleLike(ctx, i)
	case intent.UnlikePost:
		w, ok := writer.(like.Writer)
		if !ok {
			return "", fmt.Errorf("command: like writer is required")
		}
		return "", like.NewHandler(w).HandleUnlike(ctx, i)
	case intent.RepostPost:
		w, ok := writer.(repost.Writer)
		if !ok {
			return "", fmt.Errorf("command: repost writer is required")
		}
		return repost.NewHandler(w).HandleRepost(ctx, i)
	case intent.UnrepostPost:
		w, ok := writer.(repost.Writer)
		if !ok {
			return "", fmt.Errorf("command: repost writer is required")
		}
		return "", repost.NewHandler(w).HandleUnrepost(ctx, i)
	case intent.CreatePost:
		w, ok := writer.(post.Writer)
		if !ok {
			return "", fmt.Errorf("command: post writer is required")
		}
		return post.NewHandler(w).HandleCreate(ctx, i)
	case intent.UpdateContentFiltering:
		w, ok := writer.(settings.Writer)
		if !ok {
			return "", fmt.Errorf("command: settings writer is required")
		}
		_, err := settings.NewHandler(w).HandleContentFiltering(ctx, i)
		return "", err
	case intent.UpdateThreading:
		w, ok := writer.(settings.Writer)
		if !ok {
			return "", fmt.Errorf("command: settings writer is required")
		}
		_, err := settings.NewHandler(w).HandleThreading(ctx, i)
		return "", err
	default:
		return "", fmt.Errorf("command: unknown intent %T", i)
	}
}
