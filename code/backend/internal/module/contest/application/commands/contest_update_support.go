package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"
	"time"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func (s *ContestService) loadContestForUpdate(ctx context.Context, id int64) (*contestentity.Contest, error) {
	contest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrContestNotFound) {
			return nil, contestcontracts.ErrContestNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return contest, nil
}

func validateContestUpdateRequest(contest *contestentity.Contest, req UpdateContestInput, now time.Time) error {
	if req.Status != nil && *req.Status != contest.Status {
		if err := domain.ValidateStatusTransition(contest.Status, *req.Status); err != nil {
			return contestcontracts.ErrInvalidStatusTransition
		}
	}

	if err := validateContestEarlyEndOverride(contest, req, now); err != nil {
		return err
	}

	if contest.Status == contestentity.ContestStatusRegistration || contest.Status == contestentity.ContestStatusRunning || contest.Status == contestentity.ContestStatusEnded {
		if req.StartTime != nil {
			return contestcontracts.ErrContestAlreadyStarted
		}
	}

	if contest.Status == contestentity.ContestStatusRunning || contest.Status == contestentity.ContestStatusEnded {
		if req.EndTime != nil {
			return contestcontracts.ErrContestAlreadyStarted
		}
	}

	if req.Mode != nil && *req.Mode != contest.Mode && contest.Status != contestentity.ContestStatusDraft {
		return contestcontracts.ErrCannotModifyAfterDraft
	}

	return nil
}

func validateContestEarlyEndOverride(contest *contestentity.Contest, req UpdateContestInput, now time.Time) error {
	if contest == nil || req.Status == nil || *req.Status != contestentity.ContestStatusEnded || contest.Status == contestentity.ContestStatusEnded {
		return nil
	}
	if domain.ContestHasEndedAt(contest, now) {
		return nil
	}

	forced, _, err := normalizeForceOverride(req.ForceOverride, req.OverrideReason)
	if err != nil {
		return err
	}
	if forced {
		return nil
	}
	return contestcontracts.ErrContestEarlyEndRequiresOverride
}

func applyContestUpdateFields(contest *contestentity.Contest, req UpdateContestInput) error {
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
		return contestcontracts.ErrInvalidTimeRange
	}
	if req.Status != nil {
		if *req.Status != contest.Status {
			contest.StatusVersion++
		}
		contest.Status = *req.Status
	}

	return nil
}
