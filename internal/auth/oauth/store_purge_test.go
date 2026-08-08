package oauth

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	indigooauth "github.com/bluesky-social/indigo/atproto/auth/oauth"
)

func TestSQLiteStore_PurgesExpiredAuthRequests(t *testing.T) {
	t.Parallel()

	store, err := OpenSQLiteStore(filepath.Join(t.TempDir(), "oauth.db"))
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	info := indigooauth.AuthRequestData{
		State:                   "stale",
		AuthServerURL:           "https://pds.example",
		Scopes:                  []string{"atproto"},
		RequestURI:              "urn:ietf:params:oauth:request_uri:abc",
		AuthServerTokenEndpoint: "https://pds.example/oauth/token",
		PKCEVerifier:            "verifier",
		DPoPAuthServerNonce:     "nonce",
		DPoPPrivateKeyMultibase: "zKey",
	}
	if err := store.SaveAuthRequestInfo(ctx, info); err != nil {
		t.Fatalf("SaveAuthRequestInfo: %v", err)
	}

	cutoff := time.Now().Add(-2 * authRequestTTL).Unix()
	if _, err := store.db.ExecContext(ctx,
		`UPDATE oauth_auth_requests SET created_at = ? WHERE state = ?`,
		cutoff, info.State,
	); err != nil {
		t.Fatalf("backdate auth request: %v", err)
	}

	if err := store.purgeExpiredAuthRequests(ctx); err != nil {
		t.Fatalf("purgeExpiredAuthRequests: %v", err)
	}
	if _, err := store.GetAuthRequestInfo(ctx, "stale"); err == nil {
		t.Fatal("GetAuthRequestInfo after purge: want error")
	}
}
