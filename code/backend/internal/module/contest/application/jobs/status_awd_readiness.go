package jobs

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *StatusUpdater) shouldBlockAutomaticAWDStart(ctx context.Context, contest *model.Contest, nextStatus string) bool {
	if contest == nil ||
		contest.Mode != model.ContestModeAWD ||
		contest.Status != model.ContestStatusRegistration ||
		nextStatus != model.ContestStatusRunning ||
		u.awdRepo == nil {
		return false
	}

	records, err := u.awdRepo.ListReadinessChallengesByContest(ctx, contest.ID)
	if err != nil {
		u.log.Error("check_awd_auto_start_readiness_failed", zap.Int64("contest_id", contest.ID), zap.Error(err))
		return true
	}
	summary := contestdomain.BuildAWDReadiness(contest.ID, mapStatusAWDReadinessRecords(records))
	if summary.Ready {
		return false
	}
	u.log.Warn(
		"block_awd_auto_start_due_to_readiness",
		zap.Int64("contest_id", contest.ID),
		zap.Int("blocking_count", summary.BlockingCount),
		zap.Strings("global_blocking_reasons", summary.GlobalBlockingReasons),
	)
	return true
}

func mapStatusAWDReadinessRecords(records []contestports.AWDReadinessChallengeRecord) []contestdomain.AWDReadinessChallenge {
	challenges := make([]contestdomain.AWDReadinessChallenge, 0, len(records))
	for _, record := range records {
		challenges = append(challenges, contestdomain.AWDReadinessChallenge{
			ServiceID:         record.ServiceID,
			AWDChallengeID:    record.AWDChallengeID,
			Title:             record.Title,
			CheckerType:       record.CheckerType,
			CheckerConfig:     record.CheckerConfig,
			ValidationState:   record.ValidationState,
			LastPreviewAt:     record.LastPreviewAt,
			LastPreviewResult: record.LastPreviewResult,
		})
	}
	return challenges
}
