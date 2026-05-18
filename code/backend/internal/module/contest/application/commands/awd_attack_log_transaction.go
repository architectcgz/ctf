package commands

import (
	"context"
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *AWDService) persistAttackLogAndScores(ctx context.Context, contestID, roundID int64, req CreateAttackLogInput, logRecord *contestentity.AWDAttackLog) error {
	now := time.Now().UTC()
	return s.repo.WithinAttackLogTransaction(ctx, func(txRepo contestports.AWDAttackLogTxRepository) error {
		if err := txRepo.CreateAttackLog(ctx, logRecord); err != nil {
			return err
		}
		if req.IsSuccess {
			if err := txRepo.ApplyAttackImpactToVictimService(
				ctx,
				roundID,
				req.VictimTeamID,
				logRecord.ServiceID,
				logRecord.AWDChallengeID,
				logRecord.ScoreGained,
				now,
			); err != nil {
				return err
			}
		}
		return txRepo.RecalculateContestTeamScores(ctx, contestID)
	})
}
