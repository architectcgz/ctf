package queries

import "time"

type AWDRoundResult struct {
	ID           int64
	ContestID    int64
	RoundNumber  int
	Status       string
	StartedAt    *time.Time
	EndedAt      *time.Time
	AttackScore  int
	DefenseScore int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
