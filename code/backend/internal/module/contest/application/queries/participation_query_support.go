package queries

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func (s *ParticipationService) ensureContestExists(ctx context.Context, contestID int64) error {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return contestcontracts.ErrContestNotFound
		}
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}
