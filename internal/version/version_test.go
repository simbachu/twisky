package version_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/version"
)

func TestShortID_TruncatesLongBuildID(t *testing.T) {
	t.Parallel()

	prev := version.BuildID
	t.Cleanup(func() { version.BuildID = prev })

	version.BuildID = "abcdef1234567890"
	got := version.ShortID()
	if got != "abcdef1" {
		t.Fatalf("ShortID() = %q, want %q", got, "abcdef1")
	}
}

func TestShortID_KeepsShortBuildID(t *testing.T) {
	t.Parallel()

	prev := version.BuildID
	t.Cleanup(func() { version.BuildID = prev })

	version.BuildID = "dev"
	got := version.ShortID()
	if got != "dev" {
		t.Fatalf("ShortID() = %q, want %q", got, "dev")
	}
}
