# twisky

An alternative to bsky.app

## Rationale

Bluesky posts are just texts, and they exist as text from an API. we can just serve that as html. Any reactivity can be solved with htmx.
The html page is composed with functions using gomponents.

## Local run

```bash
go generate ./static
go run ./cmd/server
```

Listens on `:8080` by default (`TWISKY_ADDR`).

`style.css` is generated from the `*.css` partials in `static/web/styles/` and is not checked in. `icons.svg` is packed from `internal/assets/icons/svg/icons-raw.svg` the same way. Regenerate with `go generate ./static` after a fresh clone or after editing any partial or the raw icon sheet.

Optional: `git config core.hooksPath .githooks` so a pre-commit hook re-packs when `icons-raw.svg` is staged.

## Hosting Twisky

I have a staging environment up for testing. The Docker compose is for that environment. Yours may not match.