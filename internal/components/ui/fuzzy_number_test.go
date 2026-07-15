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
		{name: "below 10K threshold", input: 9_999, want: "9999"},
		{name: "at 10K threshold", input: 10_000, want: "10K"},
		{name: "K suffix uses thousands", input: 15_000, want: "15K"},
		{name: "just below 1M", input: 999_999, want: "999K"},
		{name: "at 1M threshold", input: 1_000_000, want: "1M"},
		{name: "M suffix uses millions", input: 34_102_326, want: "34M"},
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
