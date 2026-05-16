package jobs

import (
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func (u *AWDRoundUpdater) calculateRoundPlan(contest *model.Contest, now time.Time) (int, int, bool) {
	if contest == nil {
		return 0, 0, false
	}
	if !contest.EndTime.After(contest.StartTime) {
		return 0, 0, false
	}
	effectiveNow := contestdomain.ContestEffectiveNow(contest, now)
	if effectiveNow.Before(contest.StartTime) {
		return 0, 0, false
	}

	duration := contest.EndTime.Sub(contest.StartTime)
	totalRounds := int((duration + u.cfg.RoundInterval - 1) / u.cfg.RoundInterval)
	if totalRounds <= 0 {
		return 0, 0, false
	}
	if !effectiveNow.Before(contest.EndTime) {
		return 0, totalRounds, true
	}

	activeRound := int(effectiveNow.Sub(contest.StartTime)/u.cfg.RoundInterval) + 1
	if activeRound > totalRounds {
		activeRound = totalRounds
	}
	return activeRound, totalRounds, true
}
