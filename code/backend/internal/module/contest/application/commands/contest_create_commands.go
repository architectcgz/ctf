package commands

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ContestService) CreateContest(ctx context.Context, req *dto.CreateContestReq) (*dto.ContestResp, error) {
	if !req.EndTime.After(req.StartTime) {
		return nil, errcode.ErrInvalidTimeRange
	}

	contest := &model.Contest{
		Title:       req.Title,
		Description: req.Description,
		Mode:        req.Mode,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      model.ContestStatusDraft,
	}

	if err := s.repo.Create(ctx, contest); err != nil {
		s.log.Error("create_contest_failed", zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.log.Info("contest_created", zap.Int64("contest_id", contest.ID))
	return domain.ContestRespFromModel(contest), nil
}
