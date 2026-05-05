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
	resp := contestQueryResponseMapperInst.ToContestResultBasePtr(contest)
	if resp == nil {
		return nil
	}
	resp.StartTime = domain.NormalizeContestTime(resp.StartTime)
	resp.EndTime = domain.NormalizeContestTime(resp.EndTime)
	resp.FreezeTime = domain.NormalizeContestTimePtr(resp.FreezeTime)
	resp.CreatedAt = domain.NormalizeContestTime(resp.CreatedAt)
	resp.UpdatedAt = domain.NormalizeContestTime(resp.UpdatedAt)
	return resp
}
