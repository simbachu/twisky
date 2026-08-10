package actor_test

import (
	"errors"
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
		wantKind   actor.Kind
		wantErr    error
	}{
		{
			name:     "valid handle",
			slug:     "bsky.app",
			wantID:   "bsky.app",
			wantKind: actor.KindHandle,
		},
		{
			name:     "valid social handle",
			slug:     "jay.bsky.team",
			wantID:   "jay.bsky.team",
			wantKind: actor.KindHandle,
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
			wantKind: actor.KindDID,
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
			wantKind: actor.KindHandle,
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
			wantKind: actor.KindHandle,
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
			wantKind: actor.KindHandle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := actor.ParseSlug(tt.slug)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParseSlug() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSlug() unexpected error: %v", err)
			}
			if got.Identifier != tt.wantID {
				t.Fatalf("ParseSlug() id = %q, want %q", got.Identifier, tt.wantID)
			}
			if got.Kind != tt.wantKind {
				t.Fatalf("ParseSlug() kind = %q, want %q", got.Kind, tt.wantKind)
			}
		})
	}
}

func TestSlug_Referent(t *testing.T) {
	t.Parallel()

	handle, err := actor.ParseSlug("alice.bsky.social")
	if err != nil {
		t.Fatalf("ParseSlug() err = %v", err)
	}
	if got := handle.Referent(); got != "handle alice.bsky.social" {
		t.Fatalf("handle.Referent() = %q, want handle alice.bsky.social", got)
	}

	did, err := actor.ParseSlug("did:plc:example")
	if err != nil {
		t.Fatalf("ParseSlug() err = %v", err)
	}
	if got := did.Referent(); got != "profile did:plc:example" {
		t.Fatalf("did.Referent() = %q, want profile did:plc:example", got)
	}
}


