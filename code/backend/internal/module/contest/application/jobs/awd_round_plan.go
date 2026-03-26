package jobs

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func (u *AWDRoundUpdater) calculateRoundPlan(contest *model.Contest, now time.Time) (int, int, bool) {
	if contest == nil {
		return 0, 0, false
	}
	if !contest.EndTime.After(contest.StartTime) {
		return 0, 0, false
	}
	if now.Before(contest.StartTime) {
		return 0, 0, false
	}

	duration := contest.EndTime.Sub(contest.StartTime)
	totalRounds := int((duration + u.cfg.RoundInterval - 1) / u.cfg.RoundInterval)
	if totalRounds <= 0 {
		return 0, 0, false
	}
	if !now.Before(contest.EndTime) {
		return 0, totalRounds, true
	}

	activeRound := int(now.Sub(contest.StartTime)/u.cfg.RoundInterval) + 1
	if activeRound > totalRounds {
		activeRound = totalRounds
	}
	return activeRound, totalRounds, true
}

func (u *AWDRoundUpdater) reconcileRounds(ctx context.Context, contest *model.Contest, activeRound, totalRounds int) error {
	scheduledRounds := totalRounds
	if activeRound > 0 {
		scheduledRounds = activeRound
	}

	return u.repo.WithinTransaction(ctx, func(txRepo contestports.AWDRepository) error {
		existingRounds, err := txRepo.ListRoundsByContest(ctx, contest.ID)
		if err != nil {
			return err
		}

		scoreByRound := make(map[int][2]int, len(existingRounds))
		lastAttackScore := defaultAWDRoundAttackScore
		lastDefenseScore := defaultAWDRoundDefenseScore
		for _, item := range existingRounds {
			scoreByRound[item.RoundNumber] = [2]int{item.AttackScore, item.DefenseScore}
			lastAttackScore = item.AttackScore
			lastDefenseScore = item.DefenseScore
		}

		for roundNumber := 1; roundNumber <= scheduledRounds; roundNumber++ {
			roundStart := contest.StartTime.Add(time.Duration(roundNumber-1) * u.cfg.RoundInterval)
			roundEnd := roundStart.Add(u.cfg.RoundInterval)
			if roundEnd.After(contest.EndTime) {
				roundEnd = contest.EndTime
			}

			status := model.AWDRoundStatusFinished
			var endedAt *time.Time
			if roundNumber == activeRound {
				status = model.AWDRoundStatusRunning
			} else {
				endedAt = &roundEnd
			}

			attackScore := lastAttackScore
			defenseScore := lastDefenseScore
			if scores, ok := scoreByRound[roundNumber]; ok {
				attackScore = scores[0]
				defenseScore = scores[1]
			} else {
				scoreByRound[roundNumber] = [2]int{attackScore, defenseScore}
			}
			lastAttackScore = attackScore
			lastDefenseScore = defenseScore

			record := &model.AWDRound{
				ContestID:    contest.ID,
				RoundNumber:  roundNumber,
				Status:       status,
				StartedAt:    &roundStart,
				EndedAt:      endedAt,
				AttackScore:  attackScore,
				DefenseScore: defenseScore,
			}
			if err := txRepo.UpsertRound(ctx, record); err != nil {
				return err
			}
		}
		return nil
	})
}

func (u *AWDRoundUpdater) acquireRoundLock(ctx context.Context, contestID int64, roundNumber int) (bool, error) {
	if u.redis == nil {
		return true, nil
	}
	return u.redis.SetNX(ctx, rediskeys.AWDRoundLockKey(contestID, roundNumber), "1", u.cfg.RoundLockTTL).Result()
}
