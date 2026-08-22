package actor

import (
	"net/url"
	"strings"
)

const invalidHandle = "handle.invalid"

// ProfileSlug returns the URL path segment for a profile link.
// Prefers a real handle; falls back to DID when the handle is missing or invalid.
func ProfileSlug(handle, did string) string {
	if isLinkableHandle(handle) {
		return handle
	}
	if strings.TrimSpace(did) != "" {
		return did
	}
	return handle
}

// ProfilePath returns the profile page path for a handle or DID.
func ProfilePath(handle, did string) string {
	slug := ProfileSlug(handle, did)
	if slug == "" {
		return "/"
	}
	return "/" + slug
}

// PostPath returns the post page path for an author handle or DID and post id.
func PostPath(handle, did, postID string) string {
	slug := ProfileSlug(handle, did)
	if slug == "" {
		return "/"
	}
	return "/" + slug + "/post/" + url.PathEscape(postID)
}

func isLinkableHandle(handle string) bool {
	handle = strings.TrimSpace(handle)
	if handle == "" || handle == invalidHandle {
		return false
	}
	_, err := ParseSlug(handle)
	return err == nil
}
