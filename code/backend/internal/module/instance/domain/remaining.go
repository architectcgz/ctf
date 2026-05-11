package domain

import "time"

// RemainingExtends returns how many extension attempts are still available.
func RemainingExtends(maxExtends int, extendCount int) int {
	remaining := maxExtends - extendCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RemainingTime returns remaining lifetime in seconds.
func RemainingTime(expiresAt, now time.Time) int64 {
	remaining := int64(expiresAt.Sub(now).Seconds())
	if remaining < 0 {
		return 0
	}
	return remaining
}
