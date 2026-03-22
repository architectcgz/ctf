package practice

import "time"

const (
	EventFlagAccepted = "practice.flag_accepted"
	EventHintUnlocked = "practice.hint_unlocked"
)

type FlagAcceptedEvent struct {
	UserID      int64
	ChallengeID int64
	Dimension   string
	Points      int
	OccurredAt  time.Time
}

type HintUnlockedEvent struct {
	UserID      int64
	ChallengeID int64
	Dimension   string
	HintLevel   int
	OccurredAt  time.Time
}
