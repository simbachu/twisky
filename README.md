# twisky

An alternative to bsky.app

## Rationale

Bluesky posts are just texts, and they exist as text from an API. we can just serve that as html. Any reactivity can be solved with htmx.
The html page is composed with functions using gomponents.

## Local run

```bash
go run ./cmd/server
```

Listens on `:8080` by default (`TWISKY_ADDR`).

## Hosting Twisky

I have a staging environment up for testing. The Docker compose is for that environment. Yours may not match.