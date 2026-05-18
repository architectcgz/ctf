package commands

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	"ctf-platform/pkg/errcode"
)

func (s *ContestService) CreateContest(ctx context.Context, req CreateContestInput) (*ContestResp, error) {
	startTime := domain.NormalizeContestTime(req.StartTime)
	endTime := domain.NormalizeContestTime(req.EndTime)
	if !endTime.After(startTime) {
		return nil, errcode.ErrInvalidTimeRange
	}

	contest := &contestentity.Contest{
		Title:       req.Title,
		Description: req.Description,
		Mode:        req.Mode,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      contestentity.ContestStatusDraft,
	}

	if err := s.repo.Create(ctx, contest); err != nil {
		s.log.Error("create_contest_failed", zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.log.Info("contest_created", zap.Int64("contest_id", contest.ID))
	return contestRespFromModel(contest), nil
}
