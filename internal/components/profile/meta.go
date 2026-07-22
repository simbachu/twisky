package profile

import (
	"fmt"
	"strconv"

	"github.com/simbachu/twisky/internal/components/page"
	profilequery "github.com/simbachu/twisky/internal/query/profile"
)

func profilePageMeta(view profilequery.ProfileView, publicBaseURL string) page.PageMeta {
	description := view.Description
	if description == "" {
		description = fmt.Sprintf(
			"%s followers · %s following · %s posts",
			strconv.Itoa(view.Followers),
			strconv.Itoa(view.Following),
			strconv.Itoa(view.Posts),
		)
	}

	return page.PageMeta{
		Title:          fmt.Sprintf("%s (@%s)", view.DisplayName, view.Handle),
		Description:    description,
		CanonicalURL:   page.AbsoluteURL(publicBaseURL, "/"+view.Handle),
		ImageURL:       view.Avatar,
		OGType:         "profile",
		LargeImageCard: false,
	}
}
