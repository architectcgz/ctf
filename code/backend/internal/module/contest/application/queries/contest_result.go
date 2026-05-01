package queries

import (
	"time"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/contest/domain"
)

type ListContestsInput struct {
	Status *string
	Page   int
	Size   int
}

type ContestResult struct {
	ID          int64
	Title       string
	Description string
	Mode        string
	StartTime   time.Time
	EndTime     time.Time
	FreezeTime  *time.Time
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func contestResultFromModel(contest *model.Contest) *ContestResult {
	if contest == nil {
		return nil
	}
	return &ContestResult{
		ID:          contest.ID,
		Title:       contest.Title,
		Description: contest.Description,
		Mode:        contest.Mode,
		StartTime:   domain.NormalizeContestTime(contest.StartTime),
		EndTime:     domain.NormalizeContestTime(contest.EndTime),
		FreezeTime:  domain.NormalizeContestTimePtr(contest.FreezeTime),
		Status:      contest.Status,
		CreatedAt:   domain.NormalizeContestTime(contest.CreatedAt),
		UpdatedAt:   domain.NormalizeContestTime(contest.UpdatedAt),
	}
}
