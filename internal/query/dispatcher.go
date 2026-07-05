package query

import (
	"context"
	"fmt"

	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/tag"
	"github.com/simbachu/twisky/internal/response"
)

// Dispatcher routes intents to their query handlers.
type Dispatcher struct {
	profile *profile.Handler
	tag     *tag.Handler
	post    *post.Handler
}

func NewDispatcher(profileHandler *profile.Handler, tagHandler *tag.Handler, postHandler *post.Handler) *Dispatcher {
	return &Dispatcher{profile: profileHandler, tag: tagHandler, post: postHandler}
}

func (d *Dispatcher) Dispatch(ctx context.Context, i intent.Intent) (response.Response, error) {
	switch i := i.(type) {
	case intent.ViewProfile:
		return d.profile.Handle(ctx, i), nil
	case intent.ViewTag:
		return d.tag.Handle(ctx, i), nil
	case intent.ViewPost:
		return d.post.Handle(ctx, i), nil
	default:
		return nil, fmt.Errorf("query: unknown intent %T", i)
	}
}
