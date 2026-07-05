package actor_test

import (
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/actor"
)

func TestParseSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		slug       string
		wantID     string
		wantKind   string
		wantErr    error
	}{
		{
			name:     "valid handle",
			slug:     "bsky.app",
			wantID:   "bsky.app",
			wantKind: "handle",
		},
		{
			name:     "valid social handle",
			slug:     "jay.bsky.team",
			wantID:   "jay.bsky.team",
			wantKind: "handle",
		},
		{
			name:    "single label is not a handle",
			slug:    "hello",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:    "empty slug",
			slug:    "",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:     "valid plc did",
			slug:     "did:plc:ar7c4by46qjdydhdevvrndac",
			wantID:   "did:plc:ar7c4by46qjdydhdevvrndac",
			wantKind: "did",
		},
		{
			name:    "malformed did",
			slug:    "did:plc:",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:     "valid handle with internal hyphen",
			slug:     "my-handle.bsky.social",
			wantID:   "my-handle.bsky.social",
			wantKind: "handle",
		},
		{
			name:    "leading dot is not a handle",
			slug:    ".bsky.app",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:    "trailing dot is not a handle",
			slug:    "bsky.app.",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:    "leading hyphen is not a handle",
			slug:    "-bsky.app",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:    "trailing hyphen is not a handle",
			slug:    "bsky.app-",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:    "consecutive dots is not a handle",
			slug:    "bsky..app",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:    "dot adjacent to hyphen is not a handle",
			slug:    "bsky.-app.com",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:    "hyphen adjacent to dot is not a handle",
			slug:    "bsky-.app.com",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:    "label longer than 63 characters is not a handle",
			slug:    strings.Repeat("a", 64) + ".com",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:    "hyphen-heavy label longer than 63 characters is not a handle",
			slug:    strings.Repeat("a-", 32) + "a.com",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:     "punycode label with doubled hyphen is a valid handle",
			slug:     "xn--ls8h.test",
			wantID:   "xn--ls8h.test",
			wantKind: "handle",
		},
		{
			name:    "numeric top-level label is not a handle",
			slug:    "example.123",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:    "single-digit top-level label is not a handle",
			slug:    "a.1",
			wantErr: actor.ErrInvalidSlug,
		},
		{
			name:     "digit-leading non-final label is a valid handle",
			slug:     "1.example.com",
			wantID:   "1.example.com",
			wantKind: "handle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotID, gotKind, err := actor.ParseSlug(tt.slug)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("ParseSlug() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSlug() unexpected error: %v", err)
			}
			if gotID != tt.wantID {
				t.Fatalf("ParseSlug() id = %q, want %q", gotID, tt.wantID)
			}
			if gotKind != tt.wantKind {
				t.Fatalf("ParseSlug() kind = %q, want %q", gotKind, tt.wantKind)
			}
		})
	}
}
