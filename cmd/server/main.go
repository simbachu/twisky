package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/command"
	twiskyhttp "github.com/simbachu/twisky/internal/http"
	"github.com/simbachu/twisky/internal/query"
	"github.com/simbachu/twisky/internal/query/post"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/tag"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	blueskyClient := bluesky.NewClient()
	_ = command.NewDispatcher() // reserved for future write intents

	queries := query.NewDispatcher(
		profile.NewHandler(blueskyClient),
		tag.NewHandler(blueskyClient),
		post.NewHandler(blueskyClient),
	)

	server := twiskyhttp.NewServer(queries)

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
