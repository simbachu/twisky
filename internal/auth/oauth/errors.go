package oauth

import (
	"errors"
	"strings"

	"github.com/bluesky-social/indigo/atproto/atclient"
)

// IsInsufficientAuth reports OAuth/API errors that indicate the session lacks permission
// or needs to be re-established (for example after scope changes).
func IsInsufficientAuth(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *atclient.APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode == 401 || apiErr.StatusCode == 403 {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "insufficient") ||
		strings.Contains(msg, "permission") ||
		strings.Contains(msg, "scope") ||
		strings.Contains(msg, "forbidden")
}
