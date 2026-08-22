package actor_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/actor"
)

func TestProfileSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		handle string
		did    string
		want   string
	}{
		{name: "handle", handle: "alice.example", did: "did:plc:alice", want: "alice.example"},
		{name: "invalid handle", handle: "handle.invalid", did: "did:plc:alice", want: "did:plc:alice"},
		{name: "empty handle", handle: "", did: "did:plc:alice", want: "did:plc:alice"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := actor.ProfileSlug(tt.handle, tt.did); got != tt.want {
				t.Fatalf("ProfileSlug() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProfilePath(t *testing.T) {
	t.Parallel()

	if got := actor.ProfilePath("alice.example", "did:plc:alice"); got != "/alice.example" {
		t.Fatalf("ProfilePath() = %q", got)
	}
	if got := actor.ProfilePath("handle.invalid", "did:plc:alice"); got != "/did:plc:alice" {
		t.Fatalf("ProfilePath() = %q", got)
	}
}

func TestPostPath(t *testing.T) {
	t.Parallel()

	if got := actor.PostPath("alice.example", "did:plc:alice", "abc/123"); got != "/alice.example/post/abc%2F123" {
		t.Fatalf("PostPath() = %q", got)
	}
}
