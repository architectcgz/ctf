package queries

import (
	"context"

	contestports "ctf-platform/internal/module/contest/ports"
	"go.uber.org/zap"

	"ctf-platform/pkg/errcode"
)

func (s *ContestService) ListContests(ctx context.Context, req ListContestsInput) ([]*ContestResult, int64, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 20
	}

	filter := buildContestListFilter(req)
	offset := (page - 1) * size
	contests, total, err := s.repo.List(ctx, filter, offset, size)
	if err != nil {
		s.log.Error("list_contests_failed", zap.Error(err))
		return nil, 0, errcode.ErrInternal.WithCause(err)
	}

	resp := make([]*ContestResult, len(contests))
	for i, c := range contests {
		resp[i] = contestResultFromModel(c)
	}
	return resp, total, nil
}

func (s *ContestService) GetContestListSummary(ctx context.Context, req ListContestsInput) (*ContestListSummaryResult, error) {
	summary, err := s.repo.Summarize(ctx, buildContestListFilter(req))
	if err != nil {
		s.log.Error("summarize_contests_failed", zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &ContestListSummaryResult{
		DraftCount:        summary.DraftCount,
		RegistrationCount: summary.RegistrationCount,
		RunningCount:      summary.RunningCount,
		FrozenCount:       summary.FrozenCount,
		EndedCount:        summary.EndedCount,
	}, nil
}

func buildContestListFilter(req ListContestsInput) contestports.ContestListFilter {
	return contestports.NewContestListFilter(req.Statuses, req.Mode, normalizeContestSort(req.SortKey, req.SortOrder))
}

func normalizeContestSort(sortKey, sortOrder string) contestports.ContestListSort {
	ascending := false
	if sortOrder == "asc" {
		ascending = true
	}

	switch sortKey {
	case "start_time":
		if ascending {
			return contestports.NewContestListSortByStartTimeAsc()
		}
		return contestports.NewContestListSortByStartTimeDesc()
	default:
		if ascending {
			return contestports.NewContestListSortByCreatedAtAsc()
		}
		return contestports.NewContestListSortByCreatedAtDesc()
	}
}
