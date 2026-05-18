package domain

import (
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
)

var validStatusTransitions = map[string][]string{
	contestentity.ContestStatusDraft:        {contestentity.ContestStatusRegistration},
	contestentity.ContestStatusRegistration: {contestentity.ContestStatusDraft, contestentity.ContestStatusRunning},
	contestentity.ContestStatusRunning:      {contestentity.ContestStatusFrozen, contestentity.ContestStatusEnded},
	contestentity.ContestStatusFrozen:       {contestentity.ContestStatusRunning, contestentity.ContestStatusEnded},
	contestentity.ContestStatusEnded:        {},
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

func IsContestImmutable(contest *contestentity.Contest) bool {
	if contest == nil {
		return false
	}
	return contest.Status == contestentity.ContestStatusRunning ||
		contest.Status == contestentity.ContestStatusFrozen ||
		contest.Status == contestentity.ContestStatusEnded
}

func IsFrozenContest(contest *contestentity.Contest, now time.Time) bool {
	if contest == nil {
		return false
	}
	if contest.Status == contestentity.ContestStatusFrozen {
		return true
	}
	if contest.FreezeTime == nil {
		return false
	}
	effectiveNow := ContestEffectiveNow(contest, now)
	return !effectiveNow.Before(contest.FreezeTime.UTC()) && effectiveNow.Before(contest.EndTime.UTC())
}

func ShouldGateAWDContestStart(mode, currentStatus string, targetStatus *string) bool {
	if targetStatus == nil {
		return false
	}
	return mode == contestentity.ContestModeAWD &&
		currentStatus != contestentity.ContestStatusRunning &&
		*targetStatus == contestentity.ContestStatusRunning
}
