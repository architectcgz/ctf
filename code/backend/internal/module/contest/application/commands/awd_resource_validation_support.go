package commands

import (
	"context"
	"errors"

	"ctf-platform/internal/apperror"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *AWDService) loadContestTeams(ctx context.Context, contestID int64) (map[int64]*contestentity.Team, error) {
	teams, err := s.repo.FindTeamsByContest(ctx, contestID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	result := make(map[int64]*contestentity.Team, len(teams))
	for _, team := range teams {
		result[team.ID] = team
	}
	return result, nil
}

func (s *AWDService) loadChallenge(ctx context.Context, challengeID int64) (*contestentity.Challenge, error) {
	challenge, err := s.repo.FindChallengeByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestAWDChallengeNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return challenge, nil
}

func (s *AWDService) resolveContestRuntimeService(ctx context.Context, contestID, serviceID int64) (*contestentity.ContestAWDService, error) {
	service, err := s.repo.FindContestAWDServiceByContestAndID(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestAWDServiceNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return service, nil
}
