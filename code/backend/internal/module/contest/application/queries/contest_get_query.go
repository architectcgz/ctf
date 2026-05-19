package queries

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/module/contest/domain"
)

func (s *ContestService) GetContest(ctx context.Context, id int64) (*ContestResult, error) {
	contest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrContestNotFound) {
			return nil, contestcontracts.ErrContestNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return contestResultFromModel(contest), nil
}
