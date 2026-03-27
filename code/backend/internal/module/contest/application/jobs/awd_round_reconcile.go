package jobs

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

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
			record, attackScore, defenseScore := u.buildReconciledRoundRecord(
				contest,
				roundNumber,
				activeRound,
				scoreByRound,
				lastAttackScore,
				lastDefenseScore,
			)
			lastAttackScore = attackScore
			lastDefenseScore = defenseScore
			if err := txRepo.UpsertRound(ctx, record); err != nil {
				return err
			}
		}
		return nil
	})
}

func (u *AWDRoundUpdater) buildReconciledRoundRecord(
	contest *model.Contest,
	roundNumber, activeRound int,
	scoreByRound map[int][2]int,
	lastAttackScore, lastDefenseScore int,
) (*model.AWDRound, int, int) {
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

	return &model.AWDRound{
		ContestID:    contest.ID,
		RoundNumber:  roundNumber,
		Status:       status,
		StartedAt:    &roundStart,
		EndedAt:      endedAt,
		AttackScore:  attackScore,
		DefenseScore: defenseScore,
	}, attackScore, defenseScore
}
