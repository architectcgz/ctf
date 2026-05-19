package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

type ContestChallengeLookupAdapter struct {
	source challengecontracts.ContestChallengeContract
}

func NewContestChallengeLookupAdapter(source challengecontracts.ContestChallengeContract) *ContestChallengeLookupAdapter {
	if source == nil {
		return nil
	}
	return &ContestChallengeLookupAdapter{source: source}
}

func (r *ContestChallengeLookupAdapter) FindByID(ctx context.Context, id int64) (*contestentity.Challenge, error) {
	challenge, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) ||
		errors.Is(err, challengeports.ErrChallengeQueryChallengeNotFound) ||
		errors.Is(err, challengeports.ErrChallengeCommandChallengeNotFound) {
		return nil, contestports.ErrContestChallengeEntityNotFound
	}
	if challenge == nil || err != nil {
		return nil, err
	}
	return &contestentity.Challenge{
		ID:         challenge.ID,
		Title:      challenge.Title,
		Category:   challenge.Category,
		Difficulty: challenge.Difficulty,
		Points:     challenge.Points,
		Status:     string(challenge.Status),
		FlagType:   challenge.FlagType,
		FlagPrefix: challenge.FlagPrefix,
	}, nil
}

func (r *ContestChallengeLookupAdapter) BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error) {
	return r.source.BatchGetSolvedStatus(ctx, userID, challengeIDs)
}

func (r *ContestChallengeLookupAdapter) BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error) {
	return r.source.BatchGetSolvedCount(ctx, challengeIDs)
}

var _ contestports.ContestChallengeCatalog = (*ContestChallengeLookupAdapter)(nil)
