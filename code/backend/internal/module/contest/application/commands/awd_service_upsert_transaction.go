package commands

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) upsertServiceCheckAndRecalculate(
	ctx context.Context,
	contestID, roundID int64,
	runtimeService *model.ContestAWDService,
	req *dto.UpsertAWDServiceCheckReq,
	checkResult string,
	defenseScore int,
	now time.Time,
) (*model.AWDTeamService, error) {
	var record *model.AWDTeamService
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
