package commands

import (
	"context"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *ChallengeService) AddChallengeToContest(ctx context.Context, contestID int64, req AddContestChallengeInput) (*ContestChallengeResp, error) {
	if _, err := s.ensureMutableContest(ctx, contestID); err != nil {
		return nil, err
	}

	challenge, err := s.challengeRepo.FindByID(ctx, req.ChallengeID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestChallengeEntityNotFound) {
			return nil, challengecontracts.ErrChallengeNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if challenge.Status != contestentity.ChallengeStatusPublished {
		return nil, contestcontracts.ErrChallengeNotPublished
	}

	exists, err := s.repo.Exists(ctx, contestID, req.ChallengeID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if exists {
		return nil, contestcontracts.ErrChallengeAlreadyAdded
	}

	points := req.Points
	if points == 0 {
		points = challenge.Points
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}

	cc := &contestentity.ContestChallenge{
		ContestID:   contestID,
		ChallengeID: req.ChallengeID,
		Points:      points,
		Order:       req.Order,
		IsVisible:   isVisible,
	}
	if err := s.repo.AddChallenge(ctx, cc); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return contestChallengeRespFromModel(cc, challenge), nil
}
