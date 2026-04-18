package commands

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *AWDService) persistAttackLogAndScores(ctx context.Context, contestID, roundID int64, req *dto.CreateAWDAttackLogReq, logRecord *model.AWDAttackLog) error {
	now := time.Now()
	return s.repo.WithinTransaction(ctx, func(txRepo contestports.AWDRepository) error {
		if err := txRepo.CreateAttackLog(ctx, logRecord); err != nil {
			return err
		}
		if req.IsSuccess {
			if err := txRepo.ApplyAttackImpactToVictimService(
				ctx,
				roundID,
				req.VictimTeamID,
				logRecord.ServiceID,
				logRecord.ChallengeID,
				logRecord.ScoreGained,
				now,
			); err != nil {
				return err
			}
		}
		return txRepo.RecalculateContestTeamScores(ctx, contestID)
	})
}
