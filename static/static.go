// Package static embeds the static web assets served by the HTTP server.
package static

import "embed"

//go:generate go run ../cmd/cssbundle

// WebFS embeds icons, scripts, and the baked stylesheet. style.css is generated
// from the *.css partials in web/styles/. Run go generate
// ./static after a fresh clone or after editing any partial.
//
//go:embed web/icons web/scripts web/styles/style.css
var WebFS embed.FS
