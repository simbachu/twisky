package oauth_test

import (
	"errors"
	"testing"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
)

func TestIsDuplicateLike(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("pds down"), false},
		{errors.New("InvalidRequest: Record already exists"), true},
		{errors.New("DuplicateCreate"), true},
	}
	for _, tc := range cases {
		if got := authoauth.IsDuplicateLike(tc.err); got != tc.want {
			t.Fatalf("IsDuplicateLike(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}
