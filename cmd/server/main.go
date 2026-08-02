package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/command"
	twiskyhttp "github.com/simbachu/twisky/internal/http"
	"github.com/simbachu/twisky/internal/moderation"
	"github.com/simbachu/twisky/internal/query"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/suggestions"
	"github.com/simbachu/twisky/internal/query/tag"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	blueskyClient := bluesky.NewClient()
	blueskyClient.SetLabelers(moderation.DefaultLabelerDIDs())
	prefs := moderation.DefaultPrefsProvider{}
	_ = command.NewDispatcher() // reserved for future write intents

	queries := query.NewDispatcher(
		profile.NewHandler(blueskyClient, prefs),
		tag.NewHandler(blueskyClient, prefs),
		post.NewHandler(blueskyClient, prefs),
	)

	publicBaseURL := envOr("TWISKY_PUBLIC_BASE_URL", "")
	auth, err := authoauth.NewService(authoauth.Config{
		PublicBaseURL: publicBaseURL,
		SessionSecret: envOr("TWISKY_SESSION_SECRET", ""),
		StorePath:     envOr("TWISKY_OAUTH_STORE_PATH", "oauth.db"),
		SecureCookies: strings.HasPrefix(strings.ToLower(publicBaseURL), "https://"),
	})
	if err != nil {
		log.Fatal(err)
	}
	if auth != nil {
		defer auth.Close()
		log.Printf("oauth enabled (client_id=%s)", auth.App.Config.ClientID)
	}

	server := twiskyhttp.NewServer(queries, suggestions.NewHandler(blueskyClient, nil), publicBaseURL, auth)

	addr := envOr("TWISKY_ADDR", ":8080")
	log.Printf("listening on %s", addr)
	if err := twiskyhttp.ListenAndServe(ctx, addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
