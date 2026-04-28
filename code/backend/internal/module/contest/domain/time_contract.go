package domain

import "time"

func NormalizeContestTime(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}
	return value.UTC()
}

func NormalizeContestTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := NormalizeContestTime(*value)
	return &normalized
}
