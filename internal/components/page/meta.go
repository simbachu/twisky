package page

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const defaultDescriptionMax = 200

// PageMeta describes HTML head metadata for a full page render.
type PageMeta struct {
	Title          string
	Description    string
	CanonicalURL   string
	ImageURL       string
	OGType         string
	LargeImageCard bool
	PublishedTime  time.Time
	AuthorURL      string
	AuthorHandle   string
	Tags           []string
	ImageAlt       string
	ImageWidth     int
	ImageHeight    int
}

// AbsoluteURL joins a public base URL with a site-relative path.
func AbsoluteURL(baseURL, path string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(baseURL, "/") + path
}

// TruncateDescription shortens plain text for embed descriptions.
func TruncateDescription(text string, max int) string {
	text = strings.TrimSpace(text)
	if max <= 0 {
		max = defaultDescriptionMax
	}
	if utf8.RuneCountInString(text) <= max {
		return text
	}
	runes := []rune(text)
	return strings.TrimSpace(string(runes[:max])) + "…"
}

func socialMetaNodes(meta PageMeta) []g.Node {
	ogType := meta.OGType
	if ogType == "" {
		ogType = "website"
	}

	nodes := []g.Node{
		Meta(
			g.Attr("property", "og:title"),
			g.Attr("content", meta.Title),
		),
		Meta(
			g.Attr("property", "og:description"),
			g.Attr("content", meta.Description),
		),
		Meta(
			g.Attr("property", "og:site_name"),
			g.Attr("content", AppName),
		),
		Meta(
			g.Attr("property", "og:type"),
			g.Attr("content", ogType),
		),
		Meta(
			g.Attr("property", "og:locale"),
			g.Attr("content", "en_US"),
		),
		Meta(
			g.Attr("name", "twitter:title"),
			g.Attr("content", meta.Title),
		),
		Meta(
			g.Attr("name", "twitter:description"),
			g.Attr("content", meta.Description),
		),
	}

	if meta.CanonicalURL != "" {
		nodes = append(nodes,
			Link(
				g.Attr("rel", "canonical"),
				g.Attr("href", meta.CanonicalURL),
			),
			Meta(
				g.Attr("property", "og:url"),
				g.Attr("content", meta.CanonicalURL),
			),
		)
	}

	if !meta.PublishedTime.IsZero() {
		nodes = append(nodes, Meta(
			g.Attr("property", "article:published_time"),
			g.Attr("content", meta.PublishedTime.UTC().Format(time.RFC3339)),
		))
	}
	if meta.AuthorURL != "" {
		nodes = append(nodes, Meta(
			g.Attr("property", "article:author"),
			g.Attr("content", meta.AuthorURL),
		))
	}
	for _, tag := range meta.Tags {
		if tag == "" {
			continue
		}
		nodes = append(nodes, Meta(
			g.Attr("property", "article:tag"),
			g.Attr("content", tag),
		))
	}
	if ogType == "profile" && meta.AuthorHandle != "" {
		nodes = append(nodes, Meta(
			g.Attr("property", "profile:username"),
			g.Attr("content", meta.AuthorHandle),
		))
	}
	if meta.AuthorHandle != "" {
		nodes = append(nodes, Meta(
			g.Attr("name", "twitter:creator"),
			g.Attr("content", "@"+meta.AuthorHandle),
		))
	}

	cardType := "summary"
	if meta.ImageURL != "" && meta.LargeImageCard {
		cardType = "summary_large_image"
	}
	nodes = append(nodes, Meta(
		g.Attr("name", "twitter:card"),
		g.Attr("content", cardType),
	))

	if meta.ImageURL != "" {
		nodes = append(nodes,
			Meta(
				g.Attr("property", "og:image"),
				g.Attr("content", meta.ImageURL),
			),
			Meta(
				g.Attr("name", "twitter:image"),
				g.Attr("content", meta.ImageURL),
			),
		)
		if meta.ImageWidth > 0 {
			nodes = append(nodes, Meta(
				g.Attr("property", "og:image:width"),
				g.Attr("content", strconv.Itoa(meta.ImageWidth)),
			))
		}
		if meta.ImageHeight > 0 {
			nodes = append(nodes, Meta(
				g.Attr("property", "og:image:height"),
				g.Attr("content", strconv.Itoa(meta.ImageHeight)),
			))
		}
		if meta.ImageAlt != "" {
			nodes = append(nodes, Meta(
				g.Attr("property", "og:image:alt"),
				g.Attr("content", meta.ImageAlt),
			))
		}
	}

	return nodes
}
