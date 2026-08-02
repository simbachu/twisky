// Package static embeds the static web assets served by the HTTP server.
package static

import "embed"

//go:generate go run ../cmd/cssbundle
//go:generate go run ../cmd/iconsprite

// WebFS embeds icons, scripts, and the baked stylesheet. style.css is generated
// from the *.css partials in web/styles/. icons.svg is packed from
// internal/assets/icons/svg/icons-raw.svg. Run go generate ./static after a
// fresh clone or after editing any partial or the raw icon sheet.
//
//go:embed web/icons web/scripts web/styles/style.css
var WebFS embed.FS
