package profile_test

import (
	"context"
	"net/http"
	"testing"

	indigoidentity "github.com/bluesky-social/indigo/atproto/identity"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/identity"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/response"
)

func TestHandler_Handle_UsesCachedDIDWhenLiveLookupFails(t *testing.T) {
	t.Parallel()

	dir := identity.NewDirectoryWithInner(indigoidentity.NewMockDirectory())
	dir.Observe("simbachu.com", "did:plc:cached")

	reader := &stubReader{
		profile: &bluesky.Profile{
			DID:    "did:plc:cached",
			Handle: "simbachu.com",
		},
		feed: &bluesky.AuthorFeedResponse{},
	}
	handler := profile.NewHandler(reader, nil, dir)

	resp := handler.Handle(context.Background(), intent.ViewProfile{
		Slug: "simbachu.com",
		Tab:  intent.ProfileTabPosts,
	})
	if _, ok := resp.(profile.ProfileView); !ok {
		t.Fatalf("Handle() type = %T, want ProfileView", resp)
	}
	if reader.lastFeedRequest.Actor != "did:plc:cached" {
		t.Fatalf("lastFeedRequest.Actor = %q, want did:plc:cached", reader.lastFeedRequest.Actor)
	}
}

func TestHandler_HandleUpstreamErrorWithCachedDIDSucceeds(t *testing.T) {
	t.Parallel()

	dir := identity.NewDirectoryWithInner(indigoidentity.NewMockDirectory())
	dir.Observe("bsky.app", "did:plc:example")

	reader := &stubReader{
		profile: &bluesky.Profile{
			DID:    "did:plc:example",
			Handle: "bsky.app",
		},
		feed: &bluesky.AuthorFeedResponse{},
	}
	handler := profile.NewHandler(reader, nil, dir)

	resp := handler.Handle(context.Background(), intent.ViewProfile{
		Slug: "bsky.app",
		Tab:  intent.ProfileTabPosts,
	})
	view, ok := resp.(profile.ProfileView)
	if !ok {
		if errResp, ok := resp.(response.ErrorResponse); ok {
			t.Fatalf("Handle() err = %q", errResp.Message)
		}
		t.Fatalf("Handle() type = %T, want ProfileView", resp)
	}
	if view.DID != "did:plc:example" {
		t.Fatalf("view.DID = %q", view.DID)
	}
}

func TestHandler_HandleNotFoundWithoutCache(t *testing.T) {
	t.Parallel()

	dir := identity.NewDirectoryWithInner(indigoidentity.NewMockDirectory())
	handler := profile.NewHandler(&stubReader{err: bluesky.ErrNotFound}, nil, dir)

	resp := handler.Handle(context.Background(), intent.ViewProfile{
		Slug: "missing.example",
		Tab:  intent.ProfileTabPosts,
	})
	errResp, ok := resp.(response.ErrorResponse)
	if !ok {
		t.Fatalf("Handle() type = %T, want ErrorResponse", resp)
	}
	if errResp.Status != http.StatusNotFound {
		t.Fatalf("Handle() status = %d, want %d", errResp.Status, http.StatusNotFound)
	}
}
