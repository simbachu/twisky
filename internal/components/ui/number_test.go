package ui_test

import (
	"testing"

	"github.com/simbachu/twisky/internal/components/ui"
)

func TestFormatGroupedNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input int
		want  string
	}{
		{name: "zero", input: 0, want: "0"},
		{name: "below threshold", input: 5, want: "5"},
		{name: "just below threshold", input: 9_999, want: "9999"},
		{name: "at threshold", input: 10_000, want: "10\u2009000"},
		{name: "five digits", input: 12_345, want: "12\u2009345"},
		{name: "seven digits", input: 1_234_567, want: "1\u2009234\u2009567"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ui.FormatGroupedNumber(tc.input); got != tc.want {
				t.Fatalf("FormatGroupedNumber(%d) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
