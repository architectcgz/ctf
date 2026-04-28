package commands

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) calculateActiveRoundNumber(contest *model.Contest, now time.Time) (int, bool) {
	if contest == nil || s.awdConfig.RoundInterval <= 0 {
		return 0, false
	}
	if !contest.EndTime.After(contest.StartTime) {
		return 0, false
	}
	if now.Before(contest.StartTime) || !now.Before(contest.EndTime) {
		return 0, false
	}

	duration := contest.EndTime.Sub(contest.StartTime)
	totalRounds := int((duration + s.awdConfig.RoundInterval - 1) / s.awdConfig.RoundInterval)
	if totalRounds <= 0 {
		return 0, false
	}

	activeRound := int(now.Sub(contest.StartTime)/s.awdConfig.RoundInterval) + 1
	if activeRound > totalRounds {
		activeRound = totalRounds
	}
	return activeRound, true
}

func (s *AWDService) resolveCurrentRoundID(ctx context.Context, contestID int64) (int64, error) {
	if !s.isLiveContestWindow(ctx, contestID) {
		return 0, nil
	}
	round, err := s.resolveCurrentRound(ctx, contestID)
	if err != nil {
		if err == errcode.ErrAWDRoundNotActive {
			return 0, nil
		}
		return 0, err
	}
	return round.ID, nil
}

func (s *AWDService) isLiveContestWindow(ctx context.Context, contestID int64) bool {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil || contest == nil {
		return false
	}
	now := time.Now().UTC()
	if !now.Before(contest.EndTime) {
		return false
	}
	return contest.Status == model.ContestStatusRunning || contest.Status == model.ContestStatusFrozen
}
