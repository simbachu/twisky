package identity_test

import (
	"context"
	"errors"
	"testing"

	indigoidentity "github.com/bluesky-social/indigo/atproto/identity"
	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/identity"
)

type stubProfileReader struct {
	profile *bluesky.Profile
	err     error
	calls   int
}

func (s *stubProfileReader) GetProfile(_ context.Context, _ string) (*bluesky.Profile, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.profile, nil
}

func TestDirectory_Resolve_DIDSlugSkipsLookup(t *testing.T) {
	t.Parallel()

	dir := identity.NewDirectory()
	reader := &stubProfileReader{profile: &bluesky.Profile{DID: "did:plc:other"}}

	did, err := dir.Resolve(context.Background(), actor.Slug{
		Identifier: "did:plc:example",
		Kind:       actor.KindDID,
	}, reader)
	if err != nil {
		t.Fatalf("Resolve() err = %v", err)
	}
	if did != "did:plc:example" {
		t.Fatalf("did = %q, want did:plc:example", did)
	}
	if reader.calls != 0 {
		t.Fatalf("GetProfile calls = %d, want 0", reader.calls)
	}
}

func TestDirectory_Resolve_UsesLastKnownGood(t *testing.T) {
	t.Parallel()

	dir := identity.NewDirectory()
	dir.Observe("alice.example", "did:plc:alice")

	reader := &stubProfileReader{err: errors.New("network down")}
	did, err := dir.Resolve(context.Background(), actor.Slug{
		Identifier: "alice.example",
		Kind:       actor.KindHandle,
	}, reader)
	if err != nil {
		t.Fatalf("Resolve() err = %v", err)
	}
	if did != "did:plc:alice" {
		t.Fatalf("did = %q, want did:plc:alice", did)
	}
}

func TestDirectory_Resolve_AppViewFallback(t *testing.T) {
	t.Parallel()

	dir := identity.NewDirectoryWithInner(indigoidentity.NewMockDirectory())
	reader := &stubProfileReader{
		profile: &bluesky.Profile{
			DID:    "did:plc:bob",
			Handle: "bob.example",
		},
	}

	did, err := dir.Resolve(context.Background(), actor.Slug{
		Identifier: "bob.example",
		Kind:       actor.KindHandle,
	}, reader)
	if err != nil {
		t.Fatalf("Resolve() err = %v", err)
	}
	if did != "did:plc:bob" {
		t.Fatalf("did = %q, want did:plc:bob", did)
	}
	if reader.calls != 1 {
		t.Fatalf("GetProfile calls = %d, want 1", reader.calls)
	}

	got, err := dir.Resolve(context.Background(), actor.Slug{
		Identifier: "bob.example",
		Kind:       actor.KindHandle,
	}, &stubProfileReader{err: errors.New("offline")})
	if err != nil {
		t.Fatalf("cached Resolve() err = %v", err)
	}
	if got != "did:plc:bob" {
		t.Fatalf("cached did = %q, want did:plc:bob", got)
	}
}

func TestDirectory_Observe_IgnoresInvalidHandle(t *testing.T) {
	t.Parallel()

	dir := identity.NewDirectory()
	dir.Observe("handle.invalid", "did:plc:bad")
	dir.Observe("alice.example", "did:plc:alice")
	dir.Observe("handle.invalid", "did:plc:overwrite")

	if got := dir.HandleFor("did:plc:bad"); got != "" {
		t.Fatalf("HandleFor invalid = %q, want empty", got)
	}
	if got := dir.HandleFor("did:plc:alice"); got != "alice.example" {
		t.Fatalf("HandleFor alice = %q, want alice.example", got)
	}
}

func TestDirectory_DisplayHandle(t *testing.T) {
	t.Parallel()

	dir := identity.NewDirectory()
	dir.Observe("alice.example", "did:plc:alice")

	if got := dir.DisplayHandle("alice.example", "did:plc:alice"); got != "alice.example" {
		t.Fatalf("DisplayHandle live = %q", got)
	}
	if got := dir.DisplayHandle("handle.invalid", "did:plc:alice"); got != "alice.example" {
		t.Fatalf("DisplayHandle cached = %q, want alice.example", got)
	}
	if got := dir.DisplayHandle("handle.invalid", "did:plc:unknown"); got != "handle.invalid" {
		t.Fatalf("DisplayHandle fallback = %q", got)
	}
}

func TestDirectory_Resolve_NotFound(t *testing.T) {
	t.Parallel()

	dir := identity.NewDirectoryWithInner(indigoidentity.NewMockDirectory())
	reader := &stubProfileReader{err: bluesky.ErrNotFound}

	_, err := dir.Resolve(context.Background(), actor.Slug{
		Identifier: "missing.example",
		Kind:       actor.KindHandle,
	}, reader)
	if !errors.Is(err, identity.ErrNotFound) {
		t.Fatalf("Resolve() err = %v, want ErrNotFound", err)
	}
}
