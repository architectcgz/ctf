package commands

import (
	"context"
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func ensureAWDReadinessGate(ctx context.Context, repo contestports.AWDRepository, contestID int64, forceOverride *bool, overrideReason *string) error {
	forced, _, err := normalizeAWDReadinessOverride(forceOverride, overrideReason)
	if err != nil {
		return err
	}
	if repo == nil {
		return errcode.ErrInternal
	}

	summary, err := loadAWDReadinessSummary(ctx, repo, contestID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if summary.Ready || forced {
		return nil
	}
	return errcode.ErrAWDReadinessBlocked
}

func loadAWDReadinessSummary(ctx context.Context, repo contestports.AWDRepository, contestID int64) (*contestdomain.AWDReadinessSummary, error) {
	records, err := repo.ListReadinessChallengesByContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	return contestdomain.BuildAWDReadiness(contestID, mapAWDReadinessChallengeRecords(records)), nil
}

func mapAWDReadinessChallengeRecords(records []contestports.AWDReadinessChallengeRecord) []contestdomain.AWDReadinessChallenge {
	challenges := make([]contestdomain.AWDReadinessChallenge, 0, len(records))
	for _, record := range records {
		challenges = append(challenges, contestdomain.AWDReadinessChallenge{
			ChallengeID:       record.ChallengeID,
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

func normalizeAWDReadinessOverride(forceOverride *bool, overrideReason *string) (bool, string, error) {
	if forceOverride == nil || !*forceOverride {
		return false, "", nil
	}

	reason := ""
	if overrideReason != nil {
		reason = strings.TrimSpace(*overrideReason)
	}
	if reason == "" {
		return false, "", errcode.ErrInvalidParams
	}
	return true, reason, nil
}

func shouldGateAWDContestStart(contest *model.Contest, req *dto.UpdateContestReq) bool {
	if contest == nil || req == nil || req.Status == nil {
		return false
	}
	return contest.Mode == model.ContestModeAWD &&
		contest.Status != model.ContestStatusRunning &&
		*req.Status == model.ContestStatusRunning
}
