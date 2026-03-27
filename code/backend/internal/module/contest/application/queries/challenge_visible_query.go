package queries

import (
	"context"
	"errors"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ChallengeService) GetContestChallenges(ctx context.Context, userID, contestID int64) ([]*dto.ContestChallengeInfo, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestChallengeVisible
	}

	challenges, err := s.repo.ListChallenges(ctx, contestID, true)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if len(challenges) == 0 {
		return []*dto.ContestChallengeInfo{}, nil
	}

	challengeIDs := make([]int64, 0, len(challenges))
	for _, item := range challenges {
		challengeIDs = append(challengeIDs, item.ChallengeID)
	}

	solvedMap, err := s.challengeRepo.BatchGetSolvedStatus(userID, challengeIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	solvedCountMap, err := s.challengeRepo.BatchGetSolvedCount(challengeIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.ContestChallengeInfo, 0, len(challenges))
	for _, item := range challenges {
		challenge, findErr := s.challengeRepo.FindByID(item.ChallengeID)
		if findErr != nil {
			return nil, errcode.ErrInternal.WithCause(findErr)
		}
		result = append(result, &dto.ContestChallengeInfo{
			ID:          item.ID,
			ChallengeID: item.ChallengeID,
			Title:       challenge.Title,
			Category:    challenge.Category,
			Difficulty:  challenge.Difficulty,
			Points:      item.Points,
			Order:       item.Order,
			SolvedCount: solvedCountMap[item.ChallengeID],
			IsSolved:    solvedMap[item.ChallengeID],
		})
	}
	return result, nil
}
