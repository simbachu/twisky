package oauth_test

import (
	"errors"
	"testing"

	"github.com/bluesky-social/indigo/atproto/atclient"
	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
)

func TestIsInsufficientAuth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("network down"), false},
		{&atclient.APIError{StatusCode: 401, Name: "Unauthorized"}, true},
		{&atclient.APIError{StatusCode: 403, Name: "Forbidden", Message: "Insufficient scope"}, true},
		{&atclient.APIError{StatusCode: 502}, false},
		{errors.New("API request failed (HTTP 403): Forbidden"), true},
	}
	for _, tc := range cases {
		if got := authoauth.IsInsufficientAuth(tc.err); got != tc.want {
			t.Fatalf("IsInsufficientAuth(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}
