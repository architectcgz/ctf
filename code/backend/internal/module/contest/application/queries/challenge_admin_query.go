package queries

import (
	"context"
	"errors"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ChallengeService) ListAdminChallenges(ctx context.Context, contestID int64) ([]*dto.ContestChallengeResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	challenges, err := s.repo.ListChallenges(ctx, contestID, false)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.ContestChallengeResp, len(challenges))
	for i, item := range challenges {
		challenge, findErr := s.challengeRepo.FindByID(ctx, item.ChallengeID)
		if findErr != nil {
			return nil, errcode.ErrInternal.WithCause(findErr)
		}
		result[i] = contestdomain.ContestChallengeRespFromModel(item, challenge)
	}
	return result, nil
}
