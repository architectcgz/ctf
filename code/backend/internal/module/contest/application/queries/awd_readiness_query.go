package queries

import (
	"context"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) GetReadiness(ctx context.Context, contestID int64) (*AWDReadinessResult, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	records, err := s.repo.ListReadinessChallengesByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	summary := contestdomain.BuildAWDReadiness(contestID, mapAWDReadinessChallenges(records))
	items := make([]AWDReadinessItem, 0, len(summary.Items))
	for _, item := range summary.Items {
		items = append(items, AWDReadinessItem{
			ServiceID:       item.ServiceID,
			AWDChallengeID:  item.AWDChallengeID,
			Title:           item.Title,
			CheckerType:     item.CheckerType,
			ValidationState: item.ValidationState,
			LastPreviewAt:   item.LastPreviewAt,
			LastAccessURL:   item.LastAccessURL,
			BlockingReason:  item.BlockingReason,
		})
	}

	return &AWDReadinessResult{
		ContestID:                summary.ContestID,
		Ready:                    summary.Ready,
		TotalChallenges:          summary.TotalChallenges,
		PassedChallenges:         summary.PassedChallenges,
		PendingChallenges:        summary.PendingChallenges,
		FailedChallenges:         summary.FailedChallenges,
		StaleChallenges:          summary.StaleChallenges,
		MissingCheckerChallenges: summary.MissingCheckerChallenges,
		BlockingCount:            summary.BlockingCount,
		BlockingActions:          append([]string(nil), summary.BlockingActions...),
		GlobalBlockingReasons:    append([]string(nil), summary.GlobalBlockingReasons...),
		Items:                    items,
	}, nil
}

func mapAWDReadinessChallenges(records []contestports.AWDReadinessChallengeRecord) []contestdomain.AWDReadinessChallenge {
	challenges := make([]contestdomain.AWDReadinessChallenge, 0, len(records))
	for _, record := range records {
		challenges = append(challenges, contestdomain.AWDReadinessChallenge{
			ServiceID:         record.ServiceID,
			AWDChallengeID:    record.AWDChallengeID,
			Title:             record.Title,
			CheckerType:       string(record.CheckerType),
			CheckerConfig:     record.CheckerConfig,
			ValidationState:   string(record.ValidationState),
			LastPreviewAt:     record.LastPreviewAt,
			LastPreviewResult: record.LastPreviewResult,
		})
	}
	return challenges
}
