package oauth

import (
	"fmt"
	"strings"

	indigooauth "github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/simbachu/twisky/internal/auth/session"
)

// Service owns the indigo ClientApp and browser session jar.
type Service struct {
	App     App
	Config  *indigooauth.ClientConfig
	Jar     *session.Jar
	Store   *SQLiteStore
	BaseURL string
}

// Config is used to construct a Service.
type Config struct {
	PublicBaseURL string
	SessionSecret string
	StorePath     string
	SecureCookies bool
}

// NewService opens the OAuth store and builds ClientApp + session jar.
// Returns nil, nil when SessionSecret is empty (auth disabled).
func NewService(cfg Config) (*Service, error) {
	if strings.TrimSpace(cfg.SessionSecret) == "" {
		return nil, nil
	}
	storePath := cfg.StorePath
	if storePath == "" {
		storePath = "oauth.db"
	}
	store, err := OpenSQLiteStore(storePath)
	if err != nil {
		return nil, err
	}
	clientCfg := NewConfig(cfg.PublicBaseURL, DefaultScopes)
	app := indigooauth.NewClientApp(&clientCfg, store)
	jar := session.NewJar([]byte(cfg.SessionSecret))
	jar.SetSecure(cfg.SecureCookies)
	return &Service{
		App:     app,
		Config:  app.Config,
		Jar:     jar,
		Store:   store,
		BaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
	}, nil
}

func (s *Service) Close() error {
	if s == nil || s.Store == nil {
		return nil
	}
	return s.Store.Close()
}

func (s *Service) ClientMetadata() (indigooauth.ClientMetadata, error) {
	if s == nil || s.App == nil {
		return indigooauth.ClientMetadata{}, fmt.Errorf("oauth not configured")
	}
	return EnrichMetadata(s.Config, s.BaseURL)
}
