package evidence

import "time"

type Query struct {
	ChallengeID *int64
	ContestID   *int64
	RoundID     *int64
	EventType   string
	From        *time.Time
	To          *time.Time
	Limit       int
	Offset      int
}
