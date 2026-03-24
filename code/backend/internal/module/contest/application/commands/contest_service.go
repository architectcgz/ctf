package commands

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type ContestService struct {
	repo contestports.Repository
	log  *zap.Logger
}

func NewContestService(repo contestports.Repository, log *zap.Logger) *ContestService {
	if log == nil {
		log = zap.NewNop()
	}
	return &ContestService{repo: repo, log: log}
}

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

func (s *ContestService) UpdateContest(ctx context.Context, id int64, req *dto.UpdateContestReq) (*dto.ContestResp, error) {
	contest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if req.Status != nil && *req.Status != contest.Status {
		if !domain.IsValidTransition(contest.Status, *req.Status) {
			return nil, errcode.ErrInvalidStatusTransition
		}
	}

	if contest.Status == model.ContestStatusRegistration || contest.Status == model.ContestStatusRunning || contest.Status == model.ContestStatusEnded {
		if req.StartTime != nil {
			return nil, errcode.ErrContestAlreadyStarted
		}
	}

	if contest.Status == model.ContestStatusRunning || contest.Status == model.ContestStatusEnded {
		if req.EndTime != nil {
			return nil, errcode.ErrContestAlreadyStarted
		}
	}

	if req.Mode != nil && *req.Mode != contest.Mode {
		if contest.Status != model.ContestStatusDraft {
			return nil, errcode.ErrCannotModifyAfterDraft
		}
		contest.Mode = *req.Mode
	}

	if req.Title != nil {
		contest.Title = *req.Title
	}
	if req.Description != nil {
		contest.Description = *req.Description
	}
	if req.StartTime != nil {
		contest.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		contest.EndTime = *req.EndTime
	}

	if !contest.EndTime.After(contest.StartTime) {
		return nil, errcode.ErrInvalidTimeRange
	}

	if req.Status != nil {
		contest.Status = *req.Status
	}

	if err := s.repo.Update(ctx, contest); err != nil {
		s.log.Error("update_contest_failed", zap.Int64("contest_id", id), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.log.Info("contest_updated", zap.Int64("contest_id", id))
	return domain.ContestRespFromModel(contest), nil
}
