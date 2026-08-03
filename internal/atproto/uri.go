package atproto

import (
	"errors"
	"strings"
)

const postCollection = "app.bsky.feed.post"

var ErrInvalidPostURI = errors.New("invalid post AT URI")

// PostURI is a typed at:// URI for app.bsky.feed.post records.
type PostURI struct {
	did  string
	rkey string
}

// NewPostURI builds a post AT URI from an actor DID and record key.
func NewPostURI(did, rkey string) PostURI {
	return PostURI{did: did, rkey: rkey}
}

// ParsePostURI parses a post AT URI into its DID and record key.
func ParsePostURI(raw string) (PostURI, error) {
	uri := strings.TrimSpace(raw)
	if !strings.HasPrefix(uri, "at://") {
		return PostURI{}, ErrInvalidPostURI
	}

	path := strings.TrimPrefix(uri, "at://")
	slash := strings.Index(path, "/")
	if slash <= 0 {
		return PostURI{}, ErrInvalidPostURI
	}
	did := path[:slash]
	if did == "" {
		return PostURI{}, ErrInvalidPostURI
	}

	rest := path[slash+1:]
	marker := postCollection + "/"
	idx := strings.LastIndex(rest, marker)
	if idx < 0 {
		return PostURI{}, ErrInvalidPostURI
	}
	rkey := rest[idx+len(marker):]
	if rkey == "" {
		return PostURI{}, ErrInvalidPostURI
	}
	return PostURI{did: did, rkey: rkey}, nil
}

func (p PostURI) String() string {
	return "at://" + p.did + "/" + postCollection + "/" + p.rkey
}

func (p PostURI) AuthorDID() string { return p.did }

func (p PostURI) Rkey() string { return p.rkey }

// PostRkey returns the record key from a post AT URI.
func PostRkey(uri string) (string, error) {
	parsed, err := ParsePostURI(uri)
	if err != nil {
		return "", err
	}
	return parsed.Rkey(), nil
}

// PostAuthorDID returns the author DID from a post AT URI.
func PostAuthorDID(uri string) (string, error) {
	parsed, err := ParsePostURI(uri)
	if err != nil {
		return "", err
	}
	return parsed.AuthorDID(), nil
}
