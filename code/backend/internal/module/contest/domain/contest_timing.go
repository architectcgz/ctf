package domain

import (
	"time"

	"ctf-platform/internal/model"
)

func ContestPausedDuration(contest *model.Contest) time.Duration {
	if contest == nil || contest.PausedSeconds <= 0 {
		return 0
	}
	return time.Duration(contest.PausedSeconds) * time.Second
}

func ContestEffectiveNow(contest *model.Contest, now time.Time) time.Time {
	return now.UTC().Add(-ContestPausedDuration(contest))
}

func ContestHasStartedAt(contest *model.Contest, now time.Time) bool {
	if contest == nil {
		return false
	}
	return !ContestEffectiveNow(contest, now).Before(contest.StartTime.UTC())
}

func ContestHasEndedAt(contest *model.Contest, now time.Time) bool {
	if contest == nil {
		return true
	}
	return !ContestEffectiveNow(contest, now).Before(contest.EndTime.UTC())
}

func ContestEffectiveEndTime(contest *model.Contest) time.Time {
	if contest == nil {
		return time.Time{}
	}
	return contest.EndTime.UTC().Add(ContestPausedDuration(contest))
}

func ContestEffectiveFreezeTime(contest *model.Contest) *time.Time {
	if contest == nil || contest.FreezeTime == nil {
		return nil
	}
	effective := contest.FreezeTime.UTC().Add(ContestPausedDuration(contest))
	return &effective
}

func CloneContestWithEffectiveSchedule(contest *model.Contest) *model.Contest {
	if contest == nil {
		return nil
	}
	cloned := *contest
	cloned.EndTime = ContestEffectiveEndTime(contest)
	cloned.FreezeTime = ContestEffectiveFreezeTime(contest)
	return &cloned
}
