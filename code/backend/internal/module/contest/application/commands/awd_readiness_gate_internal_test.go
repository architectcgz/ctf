package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func TestEvaluateAWDReadinessGateReturnsDecisionSnapshotAndNormalizedOverride(t *testing.T) {
	db := contesttestsupport.SetupAWDTestDB(t)
	repo := contestinfra.NewAWDRepository(db)
	now := time.Now()

	contesttestsupport.CreateAWDContestFixture(t, db, 5101, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 51011, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 5101, 51011, now)
	contesttestsupport.SyncAWDContestServiceFixture(t, db, 5101, 51011, "awd-service", "legacy_probe", `{}`, 100, 0, 0, now)
	contesttestsupport.SyncAWDContestServiceReadinessFixture(
		t,
		db,
		5101,
		51011,
		model.AWDCheckerValidationStatePending,
		nil,
		"",
	)

	decision, err := evaluateAWDReadinessGate(context.Background(), repo, 5101, boolPtr(true), strPtr("  teacher drill  "))
	if err != nil {
		t.Fatalf("evaluateAWDReadinessGate() error = %v", err)
	}
	if !decision.Allowed() {
		t.Fatalf("expected decision to allow forced override")
	}
	if !decision.ForcedOverride() {
		t.Fatalf("expected forced override to be preserved")
	}
	if got := decision.OverrideReason(); got != "teacher drill" {
		t.Fatalf("expected normalized override reason, got %q", got)
	}

	summary := decision.ReadinessSummary()
	if summary == nil {
		t.Fatalf("expected readiness summary")
	}
	if snapshot := decision.BlockingSnapshot(); snapshot == nil || snapshot.BlockingCount != 1 {
		t.Fatalf("expected blocking snapshot, got %+v", snapshot)
	}
	if summary.Ready {
		t.Fatalf("expected blocking summary, got ready: %+v", summary)
	}
	if summary.BlockingCount != 1 {
		t.Fatalf("expected one blocking item, got %+v", summary)
	}
	if len(summary.Items) != 1 || summary.Items[0].BlockingReason != contestdomain.AWDReadinessBlockingReasonPendingValidation {
		t.Fatalf("unexpected readiness items: %+v", summary.Items)
	}
}

func boolPtr(v bool) *bool { return &v }

func strPtr(v string) *string { return &v }
