package jobs

import (
	"time"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func (u *StatusUpdater) calculateStatus(contest *contestentity.Contest, now time.Time) string {
	if contest.Status == contestentity.ContestStatusDraft {
		return contestentity.ContestStatusDraft
	}

	effectiveNow := contestdomain.ContestEffectiveNow(contest, now)
	if effectiveNow.Before(contest.StartTime) {
		return contestentity.ContestStatusRegistration
	}

	if !effectiveNow.Before(contest.EndTime) {
		return contestentity.ContestStatusEnded
	}

	if contest.Status == contestentity.ContestStatusFrozen && contest.FreezeTime == nil {
		return contestentity.ContestStatusFrozen
	}

	if contest.FreezeTime != nil && !effectiveNow.Before(contest.FreezeTime.UTC()) {
		return contestentity.ContestStatusFrozen
	}

	return contestentity.ContestStatusRunning
}
