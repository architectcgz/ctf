package domain

import (
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func ContestPausedDuration(contest *contestentity.Contest) time.Duration {
	if contest == nil || contest.PausedSeconds <= 0 {
		return 0
	}
	return time.Duration(contest.PausedSeconds) * time.Second
}

func ContestEffectiveNow(contest *contestentity.Contest, now time.Time) time.Time {
	return now.UTC().Add(-ContestPausedDuration(contest))
}

func ContestHasStartedAt(contest *contestentity.Contest, now time.Time) bool {
	if contest == nil {
		return false
	}
	return !ContestEffectiveNow(contest, now).Before(contest.StartTime.UTC())
}

func ContestHasEndedAt(contest *contestentity.Contest, now time.Time) bool {
	if contest == nil {
		return true
	}
	return !ContestEffectiveNow(contest, now).Before(contest.EndTime.UTC())
}

func ContestEffectiveEndTime(contest *contestentity.Contest) time.Time {
	if contest == nil {
		return time.Time{}
	}
	return contest.EndTime.UTC().Add(ContestPausedDuration(contest))
}

func ContestEffectiveFreezeTime(contest *contestentity.Contest) *time.Time {
	if contest == nil || contest.FreezeTime == nil {
		return nil
	}
	effective := contest.FreezeTime.UTC().Add(ContestPausedDuration(contest))
	return &effective
}

func CloneContestWithEffectiveSchedule(contest *contestentity.Contest) *contestentity.Contest {
	if contest == nil {
		return nil
	}
	cloned := *contest
	cloned.EndTime = ContestEffectiveEndTime(contest)
	cloned.FreezeTime = ContestEffectiveFreezeTime(contest)
	return &cloned
}
