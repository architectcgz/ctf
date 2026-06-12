package contracts

import "time"

const (
	EventFlagAccepted               = "practice.flag_accepted"
	EventFlagAcceptedPayloadVersion = 1
)

type FlagAcceptedEvent struct {
	UserID      int64     `json:"user_id"`
	ChallengeID int64     `json:"challenge_id"`
	Dimension   string    `json:"dimension"`
	Points      int       `json:"points"`
	OccurredAt  time.Time `json:"occurred_at"`
}
