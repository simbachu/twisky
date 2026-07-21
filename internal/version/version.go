// Package version exposes the build identifier injected at link time.
package version

// BuildID is set via -ldflags "-X github.com/simbachu/twisky/internal/version.BuildID=<sha>".
// Defaults to "dev" for local builds.
var BuildID = "dev"

// ShortID returns the first 7 characters of BuildID (or the whole string if shorter).
func ShortID() string {
	if len(BuildID) <= 7 {
		return BuildID
	}
	return BuildID[:7]
}
