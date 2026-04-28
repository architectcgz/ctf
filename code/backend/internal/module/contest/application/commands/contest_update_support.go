package commands

import (
	"context"
	"errors"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ContestService) loadContestForUpdate(ctx context.Context, id int64) (*model.Contest, error) {
	contest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return contest, nil
}

func validateContestUpdateRequest(contest *model.Contest, req *dto.UpdateContestReq) error {
	if req.Status != nil && *req.Status != contest.Status {
		if !domain.IsValidTransition(contest.Status, *req.Status) {
			return errcode.ErrInvalidStatusTransition
		}
	}

	if contest.Status == model.ContestStatusRegistration || contest.Status == model.ContestStatusRunning || contest.Status == model.ContestStatusEnded {
		if req.StartTime != nil {
			return errcode.ErrContestAlreadyStarted
		}
	}

	if contest.Status == model.ContestStatusRunning || contest.Status == model.ContestStatusEnded {
		if req.EndTime != nil {
			return errcode.ErrContestAlreadyStarted
		}
	}

	if req.Mode != nil && *req.Mode != contest.Mode && contest.Status != model.ContestStatusDraft {
		return errcode.ErrCannotModifyAfterDraft
	}

	return nil
}

func applyContestUpdateFields(contest *model.Contest, req *dto.UpdateContestReq) error {
	if req.Mode != nil && *req.Mode != contest.Mode {
		contest.Mode = *req.Mode
	}
	if req.Title != nil {
		contest.Title = *req.Title
	}
	if req.Description != nil {
		contest.Description = *req.Description
	}
	if req.StartTime != nil {
		contest.StartTime = domain.NormalizeContestTime(*req.StartTime)
	}
	if req.EndTime != nil {
		contest.EndTime = domain.NormalizeContestTime(*req.EndTime)
	}

	if !contest.EndTime.After(contest.StartTime) {
		return errcode.ErrInvalidTimeRange
	}
	if req.Status != nil {
		contest.Status = *req.Status
	}

	return nil
}
