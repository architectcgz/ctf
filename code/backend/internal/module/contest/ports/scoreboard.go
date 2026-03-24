package ports

import "time"

type ScoreboardTeamStats struct {
	SolvedCount      int
	LastSubmissionAt *time.Time
}
