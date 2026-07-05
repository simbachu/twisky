package atproto_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/atproto"
)

func TestPostRkey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		uri     string
		want    string
		wantErr bool
	}{
		{
			name: "valid post uri",
			uri:  "at://did:plc:example/app.bsky.feed.post/abc123",
			want: "abc123",
		},
		{
			name: "rkey with special chars",
			uri:  "at://did:plc:example/app.bsky.feed.post/3jzfcijpj2k2",
			want: "3jzfcijpj2k2",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := atproto.PostRkey(tt.uri)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("PostRkey() err = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("PostRkey() err = %v", err)
			}
			if got != tt.want {
				t.Fatalf("PostRkey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPostURI(t *testing.T) {
	t.Parallel()

	got := atproto.PostURI("did:plc:example", "abc123")
	want := "at://did:plc:example/app.bsky.feed.post/abc123"
	if got != want {
		t.Fatalf("PostURI() = %q, want %q", got, want)
	}
}
