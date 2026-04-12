package domain

import (
	"time"

	"ctf-platform/internal/model"
)

var validStatusTransitions = map[string][]string{
	model.ContestStatusDraft:        {model.ContestStatusRegistration},
	model.ContestStatusRegistration: {model.ContestStatusDraft, model.ContestStatusRunning},
	model.ContestStatusRunning:      {model.ContestStatusFrozen, model.ContestStatusEnded},
	model.ContestStatusFrozen:       {model.ContestStatusEnded},
	model.ContestStatusEnded:        {},
}

func IsValidTransition(from, to string) bool {
	allowed, ok := validStatusTransitions[from]
	if !ok {
		return false
	}
	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}

func IsContestImmutable(contest *model.Contest) bool {
	if contest == nil {
		return false
	}
	return contest.Status == model.ContestStatusRunning ||
		contest.Status == model.ContestStatusFrozen ||
		contest.Status == model.ContestStatusEnded
}

func IsFrozenContest(contest *model.Contest, now time.Time) bool {
	if contest == nil {
		return false
	}
	if contest.Status == model.ContestStatusFrozen {
		return true
	}
	if contest.FreezeTime == nil {
		return false
	}
	return !now.Before(*contest.FreezeTime) && now.Before(contest.EndTime)
}

func ShouldGateAWDContestStart(mode, currentStatus string, targetStatus *string) bool {
	if targetStatus == nil {
		return false
	}
	return mode == model.ContestModeAWD &&
		currentStatus != model.ContestStatusRunning &&
		*targetStatus == model.ContestStatusRunning
}
