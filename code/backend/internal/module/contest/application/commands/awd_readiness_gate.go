package commands

import (
	"context"
	"strings"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type awdReadinessGateDecision struct {
	allowed          bool
	readinessSummary *contestdomain.AWDReadinessSummary
	overrideReason   string
	forcedOverride   bool
}

func (d *awdReadinessGateDecision) Allowed() bool {
	return d != nil && d.allowed
}

func (d *awdReadinessGateDecision) ReadinessSummary() *contestdomain.AWDReadinessSummary {
	if d == nil {
		return nil
	}
	return d.readinessSummary
}

func (d *awdReadinessGateDecision) BlockingSnapshot() *contestdomain.AWDReadinessSummary {
	if d == nil || d.readinessSummary == nil || d.readinessSummary.Ready {
		return nil
	}
	return d.readinessSummary
}

func (d *awdReadinessGateDecision) OverrideReason() string {
	if d == nil {
		return ""
	}
	return d.overrideReason
}

func (d *awdReadinessGateDecision) ForcedOverride() bool {
	return d != nil && d.forcedOverride
}

func ensureAWDReadinessGate(ctx context.Context, repo contestports.AWDRepository, contestID int64, forceOverride *bool, overrideReason *string) error {
	decision, err := evaluateAWDReadinessGate(ctx, repo, contestID, forceOverride, overrideReason)
	if err != nil {
		return err
	}
	recordAWDReadinessGateDecision(ctx, decision)
	if decision.Allowed() {
		return nil
	}
	return errcode.ErrAWDReadinessBlocked
}

func evaluateAWDReadinessGate(ctx context.Context, repo contestports.AWDRepository, contestID int64, forceOverride *bool, overrideReason *string) (*awdReadinessGateDecision, error) {
	forced, normalizedReason, err := normalizeAWDReadinessOverride(forceOverride, overrideReason)
	if err != nil {
		return nil, err
	}
	if repo == nil {
		return nil, errcode.ErrInternal
	}

	summary, err := loadAWDReadinessSummary(ctx, repo, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &awdReadinessGateDecision{
		allowed:          summary.Ready || forced,
		readinessSummary: summary,
		overrideReason:   normalizedReason,
		forcedOverride:   forced,
	}, nil
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

func recordAWDReadinessGateDecision(ctx context.Context, decision *awdReadinessGateDecision) {
	trace := AWDReadinessGateTraceFromContext(ctx)
	if trace == nil || decision == nil {
		return
	}
	trace.RecordDecision(decision.Allowed())
}
