// Command cssbundle concatenates the *.css partials in static/web/styles/
// into style.css for //go:embed. Invoked by go generate ./static.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Concatenation order for partials in static/web/styles/.
var partials = []string{
	"fonts.css",
	"tokens.css",
	"base.css",
	"author.css",
	"iface.css",
	"account.css",
	"profile.css",
	"post.css",
	"post-spine.css",
	"feed.css",
	"compose.css",
	"postpage.css",
	"settings.css",
}

const styleHeader = `/* DO NOT EDIT.
 * Generated from the *.css partials in static/web/styles/.
 * Regenerate with: go generate ./static
 */

`

func main() {
	stylesDir, err := resolveStylesDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cssbundle: %v\n", err)
		os.Exit(1)
	}

	var out []byte
	out = append(out, []byte(styleHeader)...)

	for i, name := range partials {
		path := filepath.Join(stylesDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cssbundle: read %s: %v\n", path, err)
			os.Exit(1)
		}
		if i > 0 {
			out = append(out, '\n')
		}
		out = append(out, data...)
		if len(data) > 0 && data[len(data)-1] != '\n' {
			out = append(out, '\n')
		}
	}

	dest := filepath.Join(stylesDir, "style.css")
	if err := os.WriteFile(dest, out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "cssbundle: write %s: %v\n", dest, err)
		os.Exit(1)
	}
}

func resolveStylesDir() (string, error) {
	if len(os.Args) > 1 {
		return os.Args[1], nil
	}
	candidates := []string{
		filepath.Join("web", "styles"),                 // go generate from ./static
		filepath.Join("static", "web", "styles"),       // go run from module root
		filepath.Join("..", "static", "web", "styles"), // go run from cmd/cssbundle
	}
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, "tokens.css")); err == nil {
			return dir, nil
		}
	}
	return "", fmt.Errorf("could not find styles directory (tokens.css); pass path as argv[1]")
}
