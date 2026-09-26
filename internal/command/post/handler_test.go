package post_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/simbachu/twisky/internal/bluesky"
	"github.com/simbachu/twisky/internal/command/post"
	"github.com/simbachu/twisky/internal/intent"
)

func tinyPNG() []byte {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

type stubWriter struct {
	text      string
	facets    []bluesky.Facet
	reply     *intent.ReplyTo
	image     *intent.PostImage
	recordURI string
	err       error

	handles    map[string]string
	resolveErr error
}

func (s *stubWriter) CreatePost(_ context.Context, text string, facets []bluesky.Facet, reply *intent.ReplyTo, image *intent.PostImage) (string, error) {
	s.text = text
	s.facets = facets
	s.reply = reply
	s.image = image
	if s.recordURI == "" {
		s.recordURI = "at://did:plc:me/app.bsky.feed.post/abc123"
	}
	return s.recordURI, s.err
}

func (s *stubWriter) ResolveHandle(_ context.Context, handle string) (string, error) {
	if s.resolveErr != nil {
		return "", s.resolveErr
	}
	if s.handles != nil {
		if did, ok := s.handles[handle]; ok {
			return did, nil
		}
	}
	return "", errors.New("not found")
}

func TestHandler_CreatePost_Success(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	uri, err := handler.HandleCreate(context.Background(), intent.CreatePost{Text: "  hello  "})
	if err != nil {
		t.Fatalf("HandleCreate() err = %v", err)
	}
	if uri != writer.recordURI {
		t.Fatalf("uri = %q, want %q", uri, writer.recordURI)
	}
	if writer.text != "hello" {
		t.Fatalf("text = %q, want trimmed hello", writer.text)
	}
	if writer.facets != nil {
		t.Fatalf("facets = %#v, want nil", writer.facets)
	}
	if writer.reply != nil {
		t.Fatalf("reply = %#v, want nil", writer.reply)
	}
}

func TestHandler_CreatePost_EmptyRejected(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{Text: "   "})
	if err == nil {
		t.Fatal("HandleCreate() err = nil, want validation error")
	}
	if writer.text != "" {
		t.Fatalf("writer called with %q, want no call", writer.text)
	}
}

func TestHandler_CreatePost_TooLongRejected(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	text := strings.Repeat("a", post.MaxPostGraphemes+1)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{Text: text})
	if err == nil {
		t.Fatal("HandleCreate() err = nil, want validation error")
	}
	if writer.text != "" {
		t.Fatalf("writer called with text, want no call")
	}
}

func TestHandler_CreatePost_ForwardsReply(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	reply := &intent.ReplyTo{
		RootURI:   "at://did:plc:root/app.bsky.feed.post/root1",
		RootCID:   "bafyroot",
		ParentURI: "at://did:plc:parent/app.bsky.feed.post/parent1",
		ParentCID: "bafyparent",
	}
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{Text: "reply text", Reply: reply})
	if err != nil {
		t.Fatalf("HandleCreate() err = %v", err)
	}
	if writer.reply == nil || *writer.reply != *reply {
		t.Fatalf("reply = %#v, want %#v", writer.reply, reply)
	}
}

func TestHandler_CreatePost_IncompleteReplyRejected(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{
		Text:  "reply text",
		Reply: &intent.ReplyTo{ParentURI: "at://did:plc:parent/app.bsky.feed.post/parent1"},
	})
	if err == nil {
		t.Fatal("HandleCreate() err = nil, want validation error")
	}
	if writer.text != "" {
		t.Fatalf("writer called with %q, want no call", writer.text)
	}
}

func TestHandler_CreatePost_AttachesLinkFacetAndShortens(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{
		Text: "see https://example.com/page",
	})
	if err != nil {
		t.Fatalf("HandleCreate() err = %v", err)
	}
	if writer.text != "see example.com/page" {
		t.Fatalf("text = %q, want shortened display", writer.text)
	}
	if len(writer.facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(writer.facets))
	}
	feature := writer.facets[0].Features[0]
	if feature.Type != "app.bsky.richtext.facet#link" || feature.URI != "https://example.com/page" {
		t.Fatalf("feature = %#v, want link to full URI", feature)
	}
}

func TestHandler_CreatePost_LongURLAllowedAfterShorten(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	// Typed text exceeds 300 graphemes only because of the long URL; shorten brings it under.
	prefix := strings.Repeat("a", 270)
	longURL := "https://example.com/" + strings.Repeat("x", 40)
	typed := prefix + " " + longURL
	if len([]rune(typed)) <= post.MaxPostGraphemes {
		t.Fatalf("typed length %d should exceed limit before shorten", len([]rune(typed)))
	}
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{Text: typed})
	if err != nil {
		t.Fatalf("HandleCreate() err = %v", err)
	}
	if len(writer.facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1 link", len(writer.facets))
	}
	if strings.Contains(writer.text, "https://") {
		t.Fatalf("text still has scheme: %q", writer.text)
	}
}

func TestHandler_CreatePost_ResolvesMentionDID(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{
		handles: map[string]string{"bsky.app": "did:plc:bsky"},
	}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{
		Text: "hi @bsky.app",
	})
	if err != nil {
		t.Fatalf("HandleCreate() err = %v", err)
	}
	if len(writer.facets) != 1 {
		t.Fatalf("len(facets) = %d, want 1", len(writer.facets))
	}
	if writer.facets[0].Features[0].DID != "did:plc:bsky" {
		t.Fatalf("did = %q, want did:plc:bsky", writer.facets[0].Features[0].DID)
	}
}

func TestHandler_CreatePost_StripsUnknownMention(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{
		Text: "hi @nobody.example",
	})
	if err != nil {
		t.Fatalf("HandleCreate() err = %v", err)
	}
	if writer.text != "hi @nobody.example" {
		t.Fatalf("text = %q, want mention text kept", writer.text)
	}
	if writer.facets != nil {
		t.Fatalf("facets = %#v, want nil after strip", writer.facets)
	}
}

func TestHandler_CreatePost_AttachesTagFacet(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{
		Text: "hello #golang",
	})
	if err != nil {
		t.Fatalf("HandleCreate() err = %v", err)
	}
	if len(writer.facets) != 1 || writer.facets[0].Features[0].Tag != "golang" {
		t.Fatalf("facets = %#v, want tag golang", writer.facets)
	}
}

func TestHandler_CreatePost_ImageOnlyAllowed(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	png := tinyPNG()
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{
		Image: &intent.PostImage{Data: png, Alt: "dot"},
	})
	if err != nil {
		t.Fatalf("HandleCreate() err = %v", err)
	}
	if writer.text != "" {
		t.Fatalf("text = %q, want empty", writer.text)
	}
	if writer.image == nil || writer.image.MIME != "image/png" {
		t.Fatalf("image = %#v, want png", writer.image)
	}
	if writer.image.Width != 1 || writer.image.Height != 1 {
		t.Fatalf("aspect = %dx%d, want 1x1", writer.image.Width, writer.image.Height)
	}
	if writer.image.Alt != "dot" {
		t.Fatalf("alt = %q, want dot", writer.image.Alt)
	}
}

func TestHandler_CreatePost_NeitherTextNorImageRejected(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{})
	if err == nil {
		t.Fatal("HandleCreate() err = nil, want validation error")
	}
	if writer.image != nil || writer.text != "" {
		t.Fatal("writer called, want no call")
	}
}

func TestHandler_CreatePost_OversizeImageRejected(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	data := make([]byte, post.MaxImageBytes+1)
	copy(data, []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{
		Text:  "ok",
		Image: &intent.PostImage{Data: data},
	})
	if err == nil {
		t.Fatal("HandleCreate() err = nil, want oversize error")
	}
	if writer.text != "" {
		t.Fatal("writer called, want no call")
	}
}

func TestHandler_CreatePost_NonImageRejected(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{
		Text:  "ok",
		Image: &intent.PostImage{Data: []byte("not an image")},
	})
	if err == nil {
		t.Fatal("HandleCreate() err = nil, want type error")
	}
	if writer.text != "" {
		t.Fatal("writer called, want no call")
	}
}

func TestHandler_CreatePost_AltTooLongRejected(t *testing.T) {
	t.Parallel()

	writer := &stubWriter{}
	handler := post.NewHandler(writer)
	_, err := handler.HandleCreate(context.Background(), intent.CreatePost{
		Text: "ok",
		Image: &intent.PostImage{
			Data: tinyPNG(),
			Alt:  strings.Repeat("a", post.MaxAltGraphemes+1),
		},
	})
	if err == nil {
		t.Fatal("HandleCreate() err = nil, want alt error")
	}
	if writer.text != "" {
		t.Fatal("writer called, want no call")
	}
}
