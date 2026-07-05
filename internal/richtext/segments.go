package richtext

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/bluesky"
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
	urlRegex            = regexp.MustCompile(`https?://\S+`)
	domainRegex         = regexp.MustCompile(`(?:^|\s)([a-z][a-z0-9]*(?:\.[a-z0-9]+)+[\S]*)`)
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
	var spans []linkSpan
	if len(facets) > 0 {
		spans = spansFromFacets(text, facets)
	} else {
		spans = spansFromRegex(text)
	}
	if len(spans) == 0 {
		return nil
	}
	return segmentsFromSpans(text, spans)
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
		if _, _, err := actor.ParseSlug(handle); err != nil {
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
	spans := make([]linkSpan, 0)

	for _, match := range urlRegex.FindAllStringIndex(text, -1) {
		if len(match) < 2 {
			continue
		}
		start, end := match[0], match[1]
		uri := trimLinkURI(string(textBytes[start:end]))
		if uri == "" {
			continue
		}
		spans = append(spans, linkSpan{
			byteStart: start,
			byteEnd:   end,
			kind:      Link,
			uri:       uri,
		})
	}

	for _, match := range domainRegex.FindAllStringSubmatchIndex(text, -1) {
		if len(match) < 4 {
			continue
		}
		domainStart := match[2]
		domainEnd := match[3]
		if domainStart < 0 || domainEnd > len(textBytes) || domainStart >= domainEnd {
			continue
		}
		if overlapsExisting(spans, domainStart, domainEnd) {
			continue
		}

		raw := string(textBytes[domainStart:domainEnd])
		raw = trailingPunctuation.ReplaceAllString(raw, "")
		if raw == "" || strings.HasPrefix(raw, "http") {
			continue
		}

		spans = append(spans, linkSpan{
			byteStart: domainStart,
			byteEnd:   domainEnd,
			kind:      Link,
			uri:       "https://" + raw,
		})
	}

	return spans
}

func overlapsExisting(spans []linkSpan, start, end int) bool {
	for _, span := range spans {
		if start < span.byteEnd && end > span.byteStart {
			return true
		}
	}
	return false
}

func trimLinkURI(raw string) string {
	raw = trailingPunctuation.ReplaceAllString(strings.TrimSpace(raw), "")
	if raw == "" {
		return ""
	}
	return raw
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
	if utf8.RuneCountInString(tag) > maxTagGraphemes {
		return false
	}
	return true
}
