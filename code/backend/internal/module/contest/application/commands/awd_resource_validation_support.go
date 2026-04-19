package commands

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) loadContestTeams(ctx context.Context, contestID int64) (map[int64]*model.Team, error) {
	teams, err := s.repo.FindTeamsByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	result := make(map[int64]*model.Team, len(teams))
	for _, team := range teams {
		result[team.ID] = team
	}
	return result, nil
}

func (s *AWDService) ensureContestChallenge(ctx context.Context, contestID, challengeID int64) error {
	ok, err := s.repo.ContestHasChallenge(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !ok {
		return errcode.ErrChallengeNotInContest
	}
	return nil
}

func (s *AWDService) loadChallenge(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	challenge, err := s.repo.FindChallengeByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return challenge, nil
}

func (s *AWDService) resolveContestRuntimeService(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
	service, err := s.repo.FindContestAWDServiceByContestAndID(ctx, contestID, serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return service, nil
}
