// Package static embeds the static web assets served by the HTTP server.
package static

import "embed"

//go:embed web
var WebFS embed.FS
