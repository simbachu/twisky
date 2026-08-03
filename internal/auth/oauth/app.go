package oauth

import (
	"context"
	"net/url"

	indigooauth "github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

// App is the indigo ClientApp surface used by HTTP handlers.
// *indigooauth.ClientApp implements App.
type App interface {
	StartAuthFlow(ctx context.Context, identifier string) (string, error)
	ProcessCallback(ctx context.Context, params url.Values) (*indigooauth.ClientSessionData, error)
	Logout(ctx context.Context, did syntax.DID, sessionID string) error
	ResumeSession(ctx context.Context, did syntax.DID, sessionID string) (*indigooauth.ClientSession, error)
}
