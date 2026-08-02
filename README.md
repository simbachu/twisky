# twisky

An alternative to bsky.app

## Rationale

Bluesky posts are just texts, and they exist as text from an API. We can just serve that as HTML. Any reactivity can be solved with htmx. The HTML page is composed with functions using gomponents.

## Local run

The server listens on `:8080` by default (`TWISKY_ADDR`).

```bash
go generate ./static
go run ./cmd/server
```

`style.css` is built from the `*.css` partials in `static/web/styles/`, and `icons.svg` is packed from `internal/assets/icons/svg/icons-raw.svg`. Neither is checked in, so run `go generate ./static` after a fresh clone or after editing those sources.

Optional: `git config core.hooksPath .githooks` so a pre-commit hook re-packs when `icons-raw.svg` is staged.

## Atproto Identity

Logging in as a user on Atproto requires authentication against the user repository via Oauth. Set `TWISKY_SESSION_SECRET` and either `TWISKY_PUBLIC_BASE_URL=https://…` (public client metadata) or leave the base URL empty for localhost client mode.
Store path for the sessions defaults to `oauth.db` via `TWISKY_OAUTH_STORE_PATH`.

## Hosting Twisky

I have a staging environment up for testing. The Docker compose is for that environment. Yours may not match.

Deploy goes through the GitHub environment `twisky_staging`. It needs vars `DEPLOY_HOST`, `DEPLOY_USER`, and `DEPLOY_PUBLIC_BASE_URL`, plus secrets `DEPLOY_SSH_PRIVATE_KEY`, `DEPLOY_SSH_KNOWN_HOSTS`, and `TWISKY_SESSION_SECRET`. The repo also needs a `GHCR_PULL_TOKEN` so the VPS can pull the image.
