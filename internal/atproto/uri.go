package atproto

import (
	"errors"
	"strings"
)

const postCollection = "app.bsky.feed.post"

var ErrInvalidPostURI = errors.New("invalid post AT URI")

// PostRkey returns the record key from a post AT URI.
// Example: at://did:plc:example/app.bsky.feed.post/abc -> abc
func PostRkey(uri string) (string, error) {
	uri = strings.TrimSpace(uri)
	if !strings.HasPrefix(uri, "at://") {
		return "", ErrInvalidPostURI
	}

	path := strings.TrimPrefix(uri, "at://")
	slash := strings.Index(path, "/")
	if slash < 0 {
		return "", ErrInvalidPostURI
	}
	path = path[slash+1:]

	marker := postCollection + "/"
	idx := strings.LastIndex(path, marker)
	if idx < 0 {
		return "", ErrInvalidPostURI
	}

	rkey := path[idx+len(marker):]
	if rkey == "" {
		return "", ErrInvalidPostURI
	}
	return rkey, nil
}

// PostURI builds a post AT URI from an actor DID and record key.
func PostURI(did, rkey string) string {
	return "at://" + did + "/" + postCollection + "/" + rkey
}
