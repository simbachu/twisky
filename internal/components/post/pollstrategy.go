package post

import "time"

const (
	newPostAge       = 5 * time.Minute
	fastPollInterval = 10 * time.Second
	slowPollInterval = 30 * time.Second

	burstPollInterval = 5 * time.Second

	repliesCooldownBase   = 20 * time.Second
	repliesCooldownAgeRef = 2 * time.Minute
	repliesCooldownMax    = 5 * time.Minute
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

// burstCountsPollInterval is the shortened counts poll gap used after a
// reply or repost count increase (hot-thread signal).
func burstCountsPollInterval() time.Duration {
	return burstPollInterval
}

// repliesFetchCooldown returns the minimum gap between GetPostThread reply
// fetches for a live post of the given age. It grows as age² from a reference
// age so fresh threads stay responsive and older ones don't hammer upstream.
func repliesFetchCooldown(age time.Duration) time.Duration {
	if age < 0 {
		age = 0
	}
	ratio := float64(age) / float64(repliesCooldownAgeRef)
	cooldown := time.Duration(float64(repliesCooldownBase) * (1 + ratio*ratio))
	if cooldown > repliesCooldownMax {
		return repliesCooldownMax
	}
	return cooldown
}
