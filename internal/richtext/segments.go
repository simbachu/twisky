package richtext

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/rivo/uniseg"
	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/bluesky"
	"golang.org/x/net/publicsuffix"
)

const (
	tagFacetType     = "app.bsky.richtext.facet#tag"
	mentionFacetType = "app.bsky.richtext.facet#mention"
	linkFacetType    = "app.bsky.richtext.facet#link"
	maxTagGraphemes  = 64
)

var (
	trailingPunctuation = regexp.MustCompile(`\p{P}+$`)
	tagRegex            = regexp.MustCompile(`(?:^|\s)([#＃])([^\s#]+)`)
	mentionRegex        = regexp.MustCompile(`(?:^|\s|\()(@[a-zA-Z0-9.-]+)`)
	// Mirrors @atproto/api URL_REGEX: start/whitespace/(, then https URL or bare domain+path.
	urlRegex = regexp.MustCompile(`(?i)(^|[\s(])((https?://\S+)|((?P<domain>[a-z][a-z0-9]*(?:\.[a-z0-9]+)+)\S*))`)
	cashtagRegex = regexp.MustCompile(`(?:^|\s|\()(\$[A-Za-z][A-Za-z0-9]{0,4})(?:\s|$|[.,;:!?)\"'\x{2019}])`)
	linkTrailingPunct = regexp.MustCompile(`[.,;:!?]$`)
)

type SegmentKind int

const (
	Plain SegmentKind = iota
	Tag
	Mention
	Link
)

type Segment struct {
	Kind    SegmentKind
	Text    string
	Tag     string
	Mention string
	URI     string
}

type linkSpan struct {
	byteStart int
	byteEnd   int
	kind      SegmentKind
	tag       string
	mention   string
	uri       string
}

func BuildSegments(text string, facets []bluesky.Facet) []Segment {
	spans := spansFromFacets(text, facets)
	if len(spans) == 0 {
		return nil
	}
	return segmentsFromSpans(text, spans)
}

// DetectFacets finds mention, link, tag, and cashtag facets using Bluesky-compatible rules.
// Mention facets use the handle as a placeholder DID until resolved at publish time.
func DetectFacets(text string) []bluesky.Facet {
	spans := spansFromRegex(text)
	if len(spans) == 0 {
		return nil
	}
	facets := make([]bluesky.Facet, 0, len(spans))
	for _, span := range spans {
		facet := bluesky.Facet{
			Index: bluesky.FacetIndex{ByteStart: span.byteStart, ByteEnd: span.byteEnd},
		}
		switch span.kind {
		case Tag:
			facet.Features = []bluesky.FacetFeature{{Type: tagFacetType, Tag: span.tag}}
		case Mention:
			facet.Features = []bluesky.FacetFeature{{Type: mentionFacetType, DID: span.mention}}
		case Link:
			facet.Features = []bluesky.FacetFeature{{Type: linkFacetType, URI: span.uri}}
		default:
			continue
		}
		facets = append(facets, facet)
	}
	return facets
}

func DetectSegments(text string) []Segment {
	return BuildSegments(text, DetectFacets(text))
}

// ShortenLinks rewrites link display text in place (scheme stripped, long paths truncated)
// while keeping each link facet's URI as the full destination. Processes last-to-first.
func ShortenLinks(text string, facets []bluesky.Facet) (string, []bluesky.Facet) {
	if len(facets) == 0 {
		return text, facets
	}
	out := []byte(text)
	result := make([]bluesky.Facet, len(facets))
	copy(result, facets)
	for i := len(result) - 1; i >= 0; i-- {
		facet := &result[i]
		if !isLinkFacet(*facet) {
			continue
		}
		start, end := facet.Index.ByteStart, facet.Index.ByteEnd
		if start < 0 || end > len(out) || start >= end {
			continue
		}
		display := string(out[start:end])
		short := toShortURL(display)
		if short == display {
			continue
		}
		shortBytes := []byte(short)
		delta := len(shortBytes) - (end - start)
		out = append(out[:start], append(shortBytes, out[end:]...)...)
		facet.Index.ByteEnd = start + len(shortBytes)
		if delta == 0 {
			continue
		}
		for j := range result {
			if j == i {
				continue
			}
			if result[j].Index.ByteStart >= end {
				result[j].Index.ByteStart += delta
				result[j].Index.ByteEnd += delta
			}
		}
	}
	return string(out), result
}

// StripInvalidMentions drops mention facets whose DID is empty or still a bare handle.
func StripInvalidMentions(facets []bluesky.Facet) []bluesky.Facet {
	if len(facets) == 0 {
		return facets
	}
	kept := make([]bluesky.Facet, 0, len(facets))
	for _, facet := range facets {
		if mention, ok := mentionDID(facet); ok {
			if mention == "" || !strings.HasPrefix(mention, "did:") {
				continue
			}
		}
		kept = append(kept, facet)
	}
	if len(kept) == 0 {
		return nil
	}
	return kept
}

func spansFromFacets(text string, facets []bluesky.Facet) []linkSpan {
	textBytes := []byte(text)
	spans := make([]linkSpan, 0, len(facets))
	for _, facet := range facets {
		span, ok := spanFromFacet(textBytes, facet)
		if !ok {
			continue
		}
		spans = append(spans, span)
	}
	if len(spans) == 0 {
		return nil
	}
	return sortAndDedupeSpans(spans)
}

func spanFromFacet(textBytes []byte, facet bluesky.Facet) (linkSpan, bool) {
	start := facet.Index.ByteStart
	end := facet.Index.ByteEnd
	if start < 0 || end > len(textBytes) || start >= end {
		return linkSpan{}, false
	}

	for _, feature := range facet.Features {
		switch feature.Type {
		case tagFacetType:
			if !validTag(feature.Tag) {
				return linkSpan{}, false
			}
			return linkSpan{
				byteStart: start,
				byteEnd:   end,
				kind:      Tag,
				tag:       feature.Tag,
			}, true
		case mentionFacetType:
			if feature.DID == "" {
				return linkSpan{}, false
			}
			return linkSpan{
				byteStart: start,
				byteEnd:   end,
				kind:      Mention,
				mention:   feature.DID,
			}, true
		case linkFacetType:
			if feature.URI == "" {
				return linkSpan{}, false
			}
			return linkSpan{
				byteStart: start,
				byteEnd:   end,
				kind:      Link,
				uri:       feature.URI,
			}, true
		}
	}
	return linkSpan{}, false
}

func spansFromRegex(text string) []linkSpan {
	textBytes := []byte(text)
	spans := make([]linkSpan, 0)
	spans = append(spans, tagSpansFromRegex(text, textBytes)...)
	spans = append(spans, cashtagSpansFromRegex(text, textBytes)...)
	spans = append(spans, mentionSpansFromRegex(text, textBytes)...)
	spans = append(spans, linkSpansFromRegex(text, textBytes)...)
	if len(spans) == 0 {
		return nil
	}
	return sortAndDedupeSpans(spans)
}

func tagSpansFromRegex(text string, textBytes []byte) []linkSpan {
	matches := tagRegex.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return nil
	}

	spans := make([]linkSpan, 0, len(matches))
	for _, match := range matches {
		if len(match) < 6 {
			continue
		}
		hashStart := match[2]
		tagBodyStart := match[4]
		tagEnd := match[5]
		if hashStart < 0 || tagEnd > len(textBytes) || hashStart >= tagEnd {
			continue
		}

		tagBody := string(textBytes[tagBodyStart:tagEnd])
		tagBody = strings.TrimSpace(tagBody)
		tagBody = trailingPunctuation.ReplaceAllString(tagBody, "")
		if !validTag(tagBody) {
			continue
		}

		spans = append(spans, linkSpan{
			byteStart: hashStart,
			byteEnd:   tagEnd,
			kind:      Tag,
			tag:       tagBody,
		})
	}
	return spans
}

func cashtagSpansFromRegex(text string, textBytes []byte) []linkSpan {
	matches := cashtagRegex.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return nil
	}

	spans := make([]linkSpan, 0, len(matches))
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		start, end := match[2], match[3]
		if start < 0 || end > len(textBytes) || start >= end {
			continue
		}
		raw := string(textBytes[start:end])
		if !strings.HasPrefix(raw, "$") {
			continue
		}
		tag := "$" + strings.ToUpper(raw[1:])
		if !validTag(tag) {
			continue
		}
		spans = append(spans, linkSpan{
			byteStart: start,
			byteEnd:   end,
			kind:      Tag,
			tag:       tag,
		})
	}
	return spans
}

func mentionSpansFromRegex(text string, textBytes []byte) []linkSpan {
	matches := mentionRegex.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return nil
	}

	spans := make([]linkSpan, 0, len(matches))
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		mentionStart := match[2]
		mentionEnd := match[3]
		if mentionStart < 0 || mentionEnd > len(textBytes) || mentionStart >= mentionEnd {
			continue
		}

		display := string(textBytes[mentionStart:mentionEnd])
		if !strings.HasPrefix(display, "@") {
			continue
		}
		handle := display[1:]
		if strings.HasSuffix(strings.ToLower(handle), ".test") {
			// Official allows *.test without full handle grammar.
		} else if _, err := actor.ParseSlug(handle); err != nil {
			continue
		}

		spans = append(spans, linkSpan{
			byteStart: mentionStart,
			byteEnd:   mentionEnd,
			kind:      Mention,
			mention:   handle,
		})
	}
	return spans
}

func linkSpansFromRegex(text string, textBytes []byte) []linkSpan {
	matches := urlRegex.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return nil
	}

	domainIdx := urlRegex.SubexpIndex("domain")
	spans := make([]linkSpan, 0, len(matches))
	for _, match := range matches {
		// Groups: 0 full, 1 lead, 2 content, 3 https URL, 4 domain+path, domain named group.
		if len(match) < 6 {
			continue
		}
		start, end := match[4], match[5]
		if start < 0 || end > len(textBytes) || start >= end {
			continue
		}

		uri := string(textBytes[start:end])
		if !strings.HasPrefix(strings.ToLower(uri), "http://") && !strings.HasPrefix(strings.ToLower(uri), "https://") {
			if domainIdx < 0 {
				continue
			}
			di := domainIdx * 2
			if di+1 >= len(match) || match[di] < 0 {
				continue
			}
			domain := string(textBytes[match[di]:match[di+1]])
			if !isValidDomain(domain) {
				continue
			}
			uri = "https://" + uri
		}

		uri, end = trimLinkSpan(uri, start, end)
		if uri == "" || start >= end {
			continue
		}
		spans = append(spans, linkSpan{
			byteStart: start,
			byteEnd:   end,
			kind:      Link,
			uri:       uri,
		})
	}
	return spans
}

func trimLinkSpan(uri string, start, end int) (string, int) {
	if linkTrailingPunct.MatchString(uri) {
		uri = uri[:len(uri)-1]
		end--
	}
	if strings.HasSuffix(uri, ")") && !strings.Contains(uri, "(") {
		uri = uri[:len(uri)-1]
		end--
	}
	return uri, end
}

func isValidDomain(domain string) bool {
	domain = strings.ToLower(domain)
	if strings.HasSuffix(domain, ".test") {
		return true
	}
	suffix, icann := publicsuffix.PublicSuffix(domain)
	if !icann || suffix == "" {
		return false
	}
	return domain == suffix || strings.HasSuffix(domain, "."+suffix)
}

func toShortURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return raw
	}
	path := ""
	if parsed.Path != "/" {
		path = parsed.Path
	}
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	if parsed.Fragment != "" {
		path += "#" + parsed.Fragment
	}
	if len(path) > 15 {
		return parsed.Host + path[:13] + "..."
	}
	return parsed.Host + path
}

func isLinkFacet(facet bluesky.Facet) bool {
	for _, feature := range facet.Features {
		if feature.Type == linkFacetType {
			return true
		}
	}
	return false
}

func mentionDID(facet bluesky.Facet) (string, bool) {
	for _, feature := range facet.Features {
		if feature.Type == mentionFacetType {
			return feature.DID, true
		}
	}
	return "", false
}

func sortAndDedupeSpans(spans []linkSpan) []linkSpan {
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].byteStart == spans[j].byteStart {
			return spans[i].byteEnd < spans[j].byteEnd
		}
		return spans[i].byteStart < spans[j].byteStart
	})

	deduped := make([]linkSpan, 0, len(spans))
	cursor := 0
	for _, span := range spans {
		if span.byteStart < cursor {
			continue
		}
		deduped = append(deduped, span)
		cursor = span.byteEnd
	}
	return deduped
}

func segmentsFromSpans(text string, spans []linkSpan) []Segment {
	textBytes := []byte(text)
	segments := make([]Segment, 0, len(spans)*2+1)
	cursor := 0

	for _, span := range spans {
		if span.byteStart < cursor {
			continue
		}
		if span.byteStart > cursor {
			segments = append(segments, Segment{
				Kind: Plain,
				Text: string(textBytes[cursor:span.byteStart]),
			})
		}
		segment := Segment{
			Kind: span.kind,
			Text: string(textBytes[span.byteStart:span.byteEnd]),
		}
		switch span.kind {
		case Tag:
			segment.Tag = span.tag
		case Mention:
			segment.Mention = span.mention
		case Link:
			segment.URI = span.uri
		}
		segments = append(segments, segment)
		cursor = span.byteEnd
	}

	if cursor < len(textBytes) {
		segments = append(segments, Segment{
			Kind: Plain,
			Text: string(textBytes[cursor:]),
		})
	}
	return segments
}

func validTag(tag string) bool {
	if tag == "" {
		return false
	}
	if uniseg.GraphemeClusterCount(tag) > maxTagGraphemes {
		return false
	}
	// Cheap reject for absurdly long UTF-16-ish lengths matching official early-out.
	if utf8.RuneCountInString(tag) > maxTagGraphemes*2 {
		return false
	}
	return true
}
