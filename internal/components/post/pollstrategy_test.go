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

func TestBurstCountsPollInterval(t *testing.T) {
	t.Parallel()

	if got := burstCountsPollInterval(); got != 5*time.Second {
		t.Fatalf("burstCountsPollInterval() = %s, want 5s", got)
	}
}

func TestRepliesFetchCooldown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		age  time.Duration
		want time.Duration
	}{
		{name: "brand new post", age: 0, want: 20 * time.Second},
		{name: "at age ref", age: 2 * time.Minute, want: 40 * time.Second},
		{name: "five minutes", age: 5 * time.Minute, want: 145 * time.Second},
		{name: "capped at max", age: time.Hour, want: 5 * time.Minute},
		{name: "negative age treated as zero", age: -time.Minute, want: 20 * time.Second},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := repliesFetchCooldown(tc.age); got != tc.want {
				t.Fatalf("repliesFetchCooldown(%s) = %s, want %s", tc.age, got, tc.want)
			}
		})
	}
}
