package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func (s *ContestService) CreateContest(ctx context.Context, req CreateContestInput) (*ContestResp, error) {
	startTime := domain.NormalizeContestTime(req.StartTime)
	endTime := domain.NormalizeContestTime(req.EndTime)
	if !endTime.After(startTime) {
		return nil, contestcontracts.ErrInvalidTimeRange
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
		return nil, apperror.ErrInternal.WithCause(err)
	}

	s.log.Info("contest_created", zap.Int64("contest_id", contest.ID))
	return contestRespFromModel(contest), nil
}
