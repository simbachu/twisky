package settings

import (
	"context"
	"fmt"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/intent"
)

// Writer reads and writes the viewer's Bluesky preference array.
type Writer interface {
	GetPreferences(ctx context.Context) (bluesky.Preferences, error)
	PutPreferences(ctx context.Context, prefs bluesky.Preferences) error
}

type Handler struct {
	writer Writer
}

func NewHandler(writer Writer) *Handler {
	return &Handler{writer: writer}
}

func (h *Handler) HandleContentFiltering(ctx context.Context, i intent.UpdateContentFiltering) (bluesky.Preferences, error) {
	if h.writer == nil {
		return bluesky.Preferences{}, fmt.Errorf("settings: writer is required")
	}
	prefs, err := h.writer.GetPreferences(ctx)
	if err != nil {
		return bluesky.Preferences{}, err
	}
	labels := make(map[string]bluesky.LabelVisibility, len(bluesky.ContentFilterLabels))
	for _, identifier := range bluesky.ContentFilterLabels {
		labels[identifier] = parseLabel(i.Labels[identifier])
	}
	updated, err := prefs.WithAdultContent(i.AdultContentEnabled)
	if err != nil {
		return bluesky.Preferences{}, err
	}
	updated, err = updated.WithContentLabels(labels)
	if err != nil {
		return bluesky.Preferences{}, err
	}
	if err := h.writer.PutPreferences(ctx, updated); err != nil {
		return bluesky.Preferences{}, err
	}
	return updated, nil
}

func (h *Handler) HandleThreading(ctx context.Context, i intent.UpdateThreading) (bluesky.Preferences, error) {
	if h.writer == nil {
		return bluesky.Preferences{}, fmt.Errorf("settings: writer is required")
	}
	prefs, err := h.writer.GetPreferences(ctx)
	if err != nil {
		return bluesky.Preferences{}, err
	}
	updated, err := prefs.WithThreadView(bluesky.ThreadViewPref{
		Sort:                    i.Sort,
		PrioritizeFollowedUsers: i.PrioritizeFollowedUsers,
	})
	if err != nil {
		return bluesky.Preferences{}, err
	}
	if err := h.writer.PutPreferences(ctx, updated); err != nil {
		return bluesky.Preferences{}, err
	}
	return updated, nil
}

func parseLabel(raw string) bluesky.LabelVisibility {
	switch bluesky.LabelVisibility(raw) {
	case bluesky.LabelHide:
		return bluesky.LabelHide
	case bluesky.LabelWarn:
		return bluesky.LabelWarn
	default:
		return bluesky.LabelIgnore
	}
}
