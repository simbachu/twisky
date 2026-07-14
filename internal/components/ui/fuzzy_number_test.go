package ui_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestFormatFuzzyNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input int
		want  string
	}{
		{name: "zero", input: 0, want: "0"},
		{name: "under thousand", input: 999, want: "999"},
		{name: "exactly thousand", input: 1000, want: "1000"},
		{name: "over thousand", input: 1500, want: "1K"},
		{name: "exactly million", input: 1_000_000, want: "1000K"},
		{name: "over million", input: 34_102_326, want: "34M"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ui.FormatFuzzyNumber(tc.input); got != tc.want {
				t.Fatalf("FormatFuzzyNumber(%d) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
