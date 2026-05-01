package queries

import "time"

type ContestAWDServiceResult struct {
	ID                   int64
	ContestID            int64
	AWDChallengeID       int64
	Title                string
	Category             string
	Difficulty           string
	DisplayName          string
	Order                int
	IsVisible            bool
	ScoreConfig          map[string]any
	RuntimeConfig        map[string]any
	ValidationState      string
	LastPreviewAt        *time.Time
	LastPreviewResultRaw string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
