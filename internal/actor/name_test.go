package actor_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/actor"
)

func TestName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		displayName string
		handle      string
		want        string
	}{
		{
			name:        "uses display name when set",
			displayName: "Bluesky",
			handle:      "bsky.app",
			want:        "Bluesky",
		},
		{
			name:   "falls back to handle when display name empty",
			handle: "dev.example",
			want:   "dev.example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := actor.Name(tt.displayName, tt.handle)
			if got != tt.want {
				t.Fatalf("Name() = %q, want %q", got, tt.want)
			}
		})
	}
}
