package jobs

import (
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func (u *StatusUpdater) calculateStatus(contest *model.Contest, now time.Time) string {
	if contest.Status == model.ContestStatusDraft {
		return model.ContestStatusDraft
	}

	effectiveNow := contestdomain.ContestEffectiveNow(contest, now)
	if effectiveNow.Before(contest.StartTime) {
		return model.ContestStatusRegistration
	}

	if !effectiveNow.Before(contest.EndTime) {
		return model.ContestStatusEnded
	}

	if contest.Status == model.ContestStatusFrozen && contest.FreezeTime == nil {
		return model.ContestStatusFrozen
	}

	if contest.FreezeTime != nil && !effectiveNow.Before(contest.FreezeTime.UTC()) {
		return model.ContestStatusFrozen
	}

	return model.ContestStatusRunning
}
