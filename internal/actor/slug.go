package actor

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidSlug = errors.New("slug is not a valid handle or DID")

	// handleRegex mirrors the AT Protocol handle grammar as implemented by
	// bluesky-social/indigo (atproto/syntax/handle.go), which itself follows
	// https://atproto.com/specs/handle. The trailing label (TLD) must start
	// with a letter; other labels may start with a letter or digit; hyphens
	// may repeat internally but not lead or trail a label.
	handleRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
)

// ParseSlug classifies a URL path segment as a handle or DID.
// Single-label strings like "hello" are rejected without an API call.
func ParseSlug(raw string) (identifier string, kind string, err error) {
	slug := strings.TrimSpace(raw)
	if slug == "" {
		return "", "", ErrInvalidSlug
	}

	if strings.HasPrefix(slug, "did:") {
		if !isDID(slug) {
			return "", "", ErrInvalidSlug
		}
		return slug, "did", nil
	}

	if !isHandle(slug) {
		return "", "", ErrInvalidSlug
	}
	return slug, "handle", nil
}

func isDID(value string) bool {
	parts := strings.Split(value, ":")
	if len(parts) < 3 {
		return false
	}
	if parts[0] != "did" {
		return false
	}
	method := parts[1]
	if method != "plc" && method != "web" && method != "key" {
		return false
	}
	for _, part := range parts[2:] {
		if part == "" {
			return false
		}
	}
	return true
}

// isHandle applies the AT Protocol handle grammar (see handleRegex). The
// 253-char cap is checked separately since the regex itself has no upper
// bound on the number of labels.
func isHandle(value string) bool {
	if len(value) > 253 {
		return false
	}
	return handleRegex.MatchString(value)
}
