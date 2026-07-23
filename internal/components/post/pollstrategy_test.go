package post

import (
	"testing"
	"time"
)

func TestAutoStartLive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		age  time.Duration
		want bool
	}{
		{name: "brand new post", age: 0, want: true},
		{name: "just under five minutes", age: 5*time.Minute - time.Second, want: true},
		{name: "exactly five minutes", age: 5 * time.Minute, want: false},
		{name: "well over five minutes", age: time.Hour, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := autoStartLive(tc.age); got != tc.want {
				t.Fatalf("autoStartLive(%s) = %v, want %v", tc.age, got, tc.want)
			}
		})
	}
}

func TestCountsPollInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		age  time.Duration
		want time.Duration
	}{
		{name: "brand new post", age: 0, want: 10 * time.Second},
		{name: "just under five minutes", age: 5*time.Minute - time.Second, want: 10 * time.Second},
		{name: "exactly five minutes", age: 5 * time.Minute, want: 30 * time.Second},
		{name: "well over five minutes", age: time.Hour, want: 30 * time.Second},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := countsPollInterval(tc.age); got != tc.want {
				t.Fatalf("countsPollInterval(%s) = %s, want %s", tc.age, got, tc.want)
			}
		})
	}
}
