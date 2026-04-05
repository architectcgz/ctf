package contracts

import "time"

const (
	EventFlagAccepted = "practice.flag_accepted"
)

type FlagAcceptedEvent struct {
	UserID      int64
	ChallengeID int64
	Dimension   string
	Points      int
	OccurredAt  time.Time
}
