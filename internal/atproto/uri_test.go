package atproto_test

import (
	"errors"
	"testing"

	"github.com/simbachu/twisky/internal/atproto"
)

func TestParsePostURI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		uri     string
		wantDID string
		wantKey string
		wantErr bool
	}{
		{
			name:    "valid post uri",
			uri:     "at://did:plc:example/app.bsky.feed.post/abc123",
			wantDID: "did:plc:example",
			wantKey: "abc123",
		},
		{
			name:    "rkey with special chars",
			uri:     "at://did:plc:example/app.bsky.feed.post/3jzfcijpj2k2",
			wantDID: "did:plc:example",
			wantKey: "3jzfcijpj2k2",
		},
		{
			name:    "missing at prefix",
			uri:     "did:plc:example/app.bsky.feed.post/abc",
			wantErr: true,
		},
		{
			name:    "wrong collection",
			uri:     "at://did:plc:example/app.bsky.feed.like/abc",
			wantErr: true,
		},
		{
			name:    "empty rkey",
			uri:     "at://did:plc:example/app.bsky.feed.post/",
			wantErr: true,
		},
		{
			name:    "missing did",
			uri:     "at:///app.bsky.feed.post/abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := atproto.ParsePostURI(tt.uri)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParsePostURI() err = nil, want error")
				}
				if !errors.Is(err, atproto.ErrInvalidPostURI) {
					t.Fatalf("ParsePostURI() err = %v, want ErrInvalidPostURI", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePostURI() err = %v", err)
			}
			if got.AuthorDID() != tt.wantDID || got.Rkey() != tt.wantKey {
				t.Fatalf("ParsePostURI() = {%q %q}, want {%q %q}", got.AuthorDID(), got.Rkey(), tt.wantDID, tt.wantKey)
			}
			if got.String() != tt.uri {
				t.Fatalf("String() = %q, want %q", got.String(), tt.uri)
			}
		})
	}
}

func TestNewPostURI(t *testing.T) {
	t.Parallel()

	got := atproto.NewPostURI("did:plc:example", "abc123")
	want := "at://did:plc:example/app.bsky.feed.post/abc123"
	if got.String() != want {
		t.Fatalf("NewPostURI().String() = %q, want %q", got.String(), want)
	}
}
