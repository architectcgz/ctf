package queries

import (
	"context"
	"errors"

	"ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ContestService) GetContest(ctx context.Context, id int64) (*ContestResult, error) {
	contest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return contestResultFromModel(contest), nil
}
