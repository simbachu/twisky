package oauth

import (
	"errors"
	"net/http"

	"github.com/bluesky-social/indigo/atproto/atclient"
	"github.com/simbachu/twisky/internal/bluesky"
)

func mapAppViewReadErr(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *atclient.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return bluesky.ErrNotFound
	}
	return err
}
