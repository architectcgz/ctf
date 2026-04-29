package contracts

import "time"

const (
	EventAWDAttackAccepted = "contest.awd.attack_accepted"
)

type AWDAttackAcceptedEvent struct {
	UserID         int64
	ContestID      int64
	AWDChallengeID int64
	Dimension      string
	OccurredAt     time.Time
}
