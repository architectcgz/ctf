package jobs

import (
	"time"

	"ctf-platform/internal/model"
)

func (u *StatusUpdater) calculateStatus(contest *model.Contest, now time.Time) string {
	if contest.Status == model.ContestStatusDraft {
		return model.ContestStatusDraft
	}

	if now.Before(contest.StartTime) {
		return model.ContestStatusRegistration
	}

	if !now.Before(contest.EndTime) {
		return model.ContestStatusEnded
	}

	if contest.FreezeTime != nil && !now.Before(*contest.FreezeTime) {
		return model.ContestStatusFrozen
	}

	return model.ContestStatusRunning
}
