package queries

import (
	"context"
	"errors"

	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ParticipationService) ensureContestExists(ctx context.Context, contestID int64) error {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}
