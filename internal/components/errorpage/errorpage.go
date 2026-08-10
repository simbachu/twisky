package errorpage

import (
	"fmt"
	"net/http"

	"github.com/simbachu/twisky/internal/components/page"
	"github.com/simbachu/twisky/internal/components/ui"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const AlertRegionID = "page-alert"

// AlertContent renders the semantic alert body used on full pages and fragments.
func AlertContent(message string, status int) g.Node {
	return Div(
		Header(H2(g.Text(statusHeading(status)))),
		P(g.Attr("role", "alert"), g.Text(message)),
	)
}

// Page renders a full document with site chrome explaining a query failure.
func Page(message string, status int, requestPath string, suggested []ui.AuthorInfo, accounts ui.AccountMenuView, publicBaseURL string) g.Node {
	title := fmt.Sprintf("%s · %s", statusHeading(status), page.AppName)
	return page.PageWithAlert(
		page.PageMeta{
			Title:        title,
			Description:  message,
			CanonicalURL: page.AbsoluteURL(publicBaseURL, requestPath),
			Path:         requestPath,
		},
		suggested,
		accounts,
		AlertContent(message, status),
	)
}

// AlertFragment renders an out-of-band replacement for the page alert region.
func AlertFragment(message string, status int) g.Node {
	return Div(
		g.Attr("id", AlertRegionID),
		g.Attr("hx-swap-oob", "true"),
		AlertContent(message, status),
	)
}

func statusHeading(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "Bad request"
	case http.StatusNotFound:
		return "Not found"
	case http.StatusBadGateway:
		return "Temporarily unavailable"
	case http.StatusInternalServerError:
		return "Something went wrong"
	default:
		if status >= 500 {
			return "Something went wrong"
		}
		if status >= 400 {
			return "Request failed"
		}
		return "Error"
	}
}
