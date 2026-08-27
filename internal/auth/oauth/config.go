package oauth

import (
	"fmt"
	"strings"

	indigooauth "github.com/bluesky-social/indigo/atproto/auth/oauth"
)

// NewConfig builds a public (or localhost) indigo client config from the app base URL.
// Empty baseURL uses localhost development client metadata.
func NewConfig(publicBaseURL string, scopes []string) indigooauth.ClientConfig {
	if len(scopes) == 0 {
		scopes = DefaultScopes
	}
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		return indigooauth.NewLocalhostConfig("http://127.0.0.1:8080/oauth/callback", scopes)
	}
	return indigooauth.NewPublicConfig(
		base+"/oauth/client-metadata.json",
		base+"/oauth/callback",
		scopes,
	)
}

// MetadataURL returns the public client metadata document URL for baseURL.
func MetadataURL(publicBaseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		return ""
	}
	return base + "/oauth/client-metadata.json"
}

// CallbackURL returns the OAuth redirect URI for baseURL.
func CallbackURL(publicBaseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		return "http://127.0.0.1:8080/oauth/callback"
	}
	return base + "/oauth/callback"
}

func strPtr(value string) *string {
	return &value
}

// EnrichMetadata adds Twisky display fields and validates the document.
func EnrichMetadata(cfg *indigooauth.ClientConfig, publicBaseURL string) (indigooauth.ClientMetadata, error) {
	meta := cfg.ClientMetadata()
	meta.ClientName = strPtr("Twisky")
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base != "" {
		meta.ClientURI = strPtr(base)
	}
	if err := meta.Validate(cfg.ClientID); err != nil {
		return indigooauth.ClientMetadata{}, fmt.Errorf("client metadata: %w", err)
	}
	return meta, nil
}
