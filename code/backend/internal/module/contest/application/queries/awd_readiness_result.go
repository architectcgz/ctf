package queries

import "time"

type AWDReadinessResult struct {
	ContestID                int64
	Ready                    bool
	TotalChallenges          int
	PassedChallenges         int
	PendingChallenges        int
	FailedChallenges         int
	StaleChallenges          int
	MissingCheckerChallenges int
	BlockingCount            int
	BlockingActions          []string
	GlobalBlockingReasons    []string
	Items                    []AWDReadinessItem
}

type AWDReadinessItem struct {
	ServiceID       int64
	AWDChallengeID  int64
	Title           string
	CheckerType     string
	ValidationState string
	LastPreviewAt   *time.Time
	LastAccessURL   *string
	BlockingReason  string
}
