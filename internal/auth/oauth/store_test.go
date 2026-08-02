package oauth_test

import (
	"context"
	"path/filepath"
	"testing"

	indigooauth "github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/simbachu/twisky/internal/auth/oauth"
)

func TestSQLiteStore_SessionRoundTrip(t *testing.T) {
	t.Parallel()

	store, err := oauth.OpenSQLiteStore(filepath.Join(t.TempDir(), "oauth.db"))
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	did := syntax.DID("did:plc:alice")
	sess := indigooauth.ClientSessionData{
		AccountDID:              did,
		SessionID:               "sess-1",
		HostURL:                 "https://pds.example",
		AuthServerURL:           "https://pds.example",
		AuthServerTokenEndpoint: "https://pds.example/oauth/token",
		Scopes:                  []string{"atproto", "transition:generic"},
		AccessToken:             "access",
		RefreshToken:            "refresh",
		DPoPAuthServerNonce:     "n1",
		DPoPHostNonce:           "n2",
		DPoPPrivateKeyMultibase: "zPrivate",
	}
	ctx := context.Background()
	if err := store.SaveSession(ctx, sess); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}

	got, err := store.GetSession(ctx, did, "sess-1")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got.AccessToken != "access" || got.RefreshToken != "refresh" {
		t.Fatalf("tokens = %q/%q", got.AccessToken, got.RefreshToken)
	}

	sess.AccessToken = "access-2"
	if err := store.SaveSession(ctx, sess); err != nil {
		t.Fatalf("SaveSession upsert: %v", err)
	}
	got, err = store.GetSession(ctx, did, "sess-1")
	if err != nil {
		t.Fatalf("GetSession after upsert: %v", err)
	}
	if got.AccessToken != "access-2" {
		t.Fatalf("AccessToken = %q, want access-2", got.AccessToken)
	}

	if err := store.DeleteSession(ctx, did, "sess-1"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if _, err := store.GetSession(ctx, did, "sess-1"); err == nil {
		t.Fatal("GetSession after delete: want error")
	}
}

func TestSQLiteStore_AuthRequestRoundTrip(t *testing.T) {
	t.Parallel()

	store, err := oauth.OpenSQLiteStore(filepath.Join(t.TempDir(), "oauth.db"))
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	did := syntax.DID("did:plc:alice")
	info := indigooauth.AuthRequestData{
		State:                   "state-1",
		AuthServerURL:           "https://pds.example",
		AccountDID:              &did,
		Scopes:                  []string{"atproto"},
		RequestURI:              "urn:ietf:params:oauth:request_uri:abc",
		AuthServerTokenEndpoint: "https://pds.example/oauth/token",
		PKCEVerifier:            "verifier",
		DPoPAuthServerNonce:     "nonce",
		DPoPPrivateKeyMultibase: "zKey",
	}
	ctx := context.Background()
	if err := store.SaveAuthRequestInfo(ctx, info); err != nil {
		t.Fatalf("SaveAuthRequestInfo: %v", err)
	}
	if err := store.SaveAuthRequestInfo(ctx, info); err == nil {
		t.Fatal("duplicate SaveAuthRequestInfo: want error")
	}

	got, err := store.GetAuthRequestInfo(ctx, "state-1")
	if err != nil {
		t.Fatalf("GetAuthRequestInfo: %v", err)
	}
	if got.PKCEVerifier != "verifier" {
		t.Fatalf("PKCEVerifier = %q", got.PKCEVerifier)
	}

	if err := store.DeleteAuthRequestInfo(ctx, "state-1"); err != nil {
		t.Fatalf("DeleteAuthRequestInfo: %v", err)
	}
	if _, err := store.GetAuthRequestInfo(ctx, "state-1"); err == nil {
		t.Fatal("GetAuthRequestInfo after delete: want error")
	}
}

func TestNewConfig_PublicFromBaseURL(t *testing.T) {
	t.Parallel()

	cfg := oauth.NewConfig("https://dev.twisky.app", []string{"atproto", "transition:generic"})
	if cfg.ClientID != "https://dev.twisky.app/oauth/client-metadata.json" {
		t.Fatalf("ClientID = %q", cfg.ClientID)
	}
	if cfg.CallbackURL != "https://dev.twisky.app/oauth/callback" {
		t.Fatalf("CallbackURL = %q", cfg.CallbackURL)
	}
}

func TestNewConfig_LocalhostWhenBaseEmpty(t *testing.T) {
	t.Parallel()

	cfg := oauth.NewConfig("", []string{"atproto", "transition:generic"})
	if cfg.CallbackURL != "http://127.0.0.1:8080/oauth/callback" {
		t.Fatalf("CallbackURL = %q", cfg.CallbackURL)
	}
	if cfg.ClientID == "" {
		t.Fatal("ClientID is empty")
	}
}
