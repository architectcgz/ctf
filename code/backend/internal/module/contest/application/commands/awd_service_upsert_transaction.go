package commands

import (
	"context"
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) upsertServiceCheckAndRecalculate(
	ctx context.Context,
	contestID, roundID int64,
	runtimeService *contestentity.ContestAWDService,
	req UpsertServiceCheckInput,
	checkResult string,
	defenseScore int,
	now time.Time,
) (*contestentity.AWDTeamService, error) {
	var record *contestentity.AWDTeamService
	if err := s.repo.WithinServiceCheckTransaction(ctx, func(txRepo contestports.AWDServiceCheckTxRepository) error {
		var txErr error
		record, txErr = txRepo.UpsertServiceCheck(
			ctx,
			roundID,
			req.TeamID,
			runtimeService.ID,
			runtimeService.AWDChallengeID,
			req.ServiceStatus,
			checkResult,
			defenseScore,
			now,
		)
		if txErr != nil {
			return txErr
		}
		return txRepo.RecalculateContestTeamScores(ctx, contestID)
	}); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return record, nil
}
