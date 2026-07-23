package post

import "time"

const (
	newPostAge       = 5 * time.Minute
	fastPollInterval = 10 * time.Second
	slowPollInterval = 30 * time.Second
)

// autoStartLive reports whether a post should start with live counts polling
// enabled by default (without the viewer opting in via the play button or
// ?live=1).
func autoStartLive(age time.Duration) bool {
	return age < newPostAge
}

// countsPollInterval returns how often to refresh a post's engagement counts
// while live, based on the post's age.
func countsPollInterval(age time.Duration) time.Duration {
	if age < newPostAge {
		return fastPollInterval
	}
	return slowPollInterval
}
