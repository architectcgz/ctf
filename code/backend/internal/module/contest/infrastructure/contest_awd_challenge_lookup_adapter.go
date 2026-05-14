package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
)

type ContestAWDChallengeLookupAdapter struct {
	source challengeports.AWDChallengeQueryRepository
}

func NewContestAWDChallengeLookupAdapter(source challengeports.AWDChallengeQueryRepository) *ContestAWDChallengeLookupAdapter {
	if source == nil {
		return nil
	}
	return &ContestAWDChallengeLookupAdapter{source: source}
}

func (r *ContestAWDChallengeLookupAdapter) FindAWDChallengeByID(ctx context.Context, id int64) (*model.AWDChallenge, error) {
	challenge, err := r.source.FindAWDChallengeByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, challengeports.ErrAWDChallengeNotFound) {
		return nil, contestports.ErrContestAWDChallengeNotFound
	}
	return challenge, err
}

func (r *ContestAWDChallengeLookupAdapter) ListAWDChallenges(ctx context.Context, query *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error) {
	return r.source.ListAWDChallenges(ctx, query)
}

var _ challengeports.AWDChallengeQueryRepository = (*ContestAWDChallengeLookupAdapter)(nil)
