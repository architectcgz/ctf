package queries

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func newAWDQueryServiceForTest(t *testing.T) (*AWDService, *gorm.DB) {
	t.Helper()

	db := contesttestsupport.SetupAWDTestDB(t)
	return NewAWDService(
		contestinfra.NewAWDRepository(db),
		contestinfra.NewRepository(db),
	), db
}

func TestAWDQueryServiceGetReadinessCountsBlockingStates(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Now()

	createAWDReadinessContestFixture(t, db, 701, now)
	createAWDReadinessChallengeFixture(t, db, 7011, "preview-failed", now)
	createAWDReadinessChallengeFixture(t, db, 7012, "pending-service", now)
	createAWDReadinessChallengeFixture(t, db, 7013, "stale-service", now)
	createAWDReadinessChallengeFixture(t, db, 7014, "passed-service", now)
	createAWDReadinessRelationFixture(t, db, &model.ContestChallenge{
		ContestID:                   701,
		ChallengeID:                 7011,
		Points:                      100,
		IsVisible:                   true,
		AWDCheckerType:              model.AWDCheckerTypeHTTPStandard,
		AWDCheckerConfig:            `{"get_flag":{"path":"/health"}}`,
		AWDCheckerValidationState:   model.AWDCheckerValidationStateFailed,
		AWDCheckerLastPreviewAt:     &now,
		AWDCheckerLastPreviewResult: `{"service_status":"down","preview_context":{"access_url":"http://preview.internal"}}`,
		CreatedAt:                   now,
		UpdatedAt:                   now,
	})
	createAWDReadinessRelationFixture(t, db, &model.ContestChallenge{
		ContestID:                 701,
		ChallengeID:               7012,
		Points:                    100,
		IsVisible:                 true,
		AWDCheckerType:            model.AWDCheckerTypeLegacyProbe,
		AWDCheckerConfig:          `{}`,
		AWDCheckerValidationState: model.AWDCheckerValidationStatePending,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	})
	createAWDReadinessRelationFixture(t, db, &model.ContestChallenge{
		ContestID:                 701,
		ChallengeID:               7013,
		Points:                    100,
		IsVisible:                 true,
		AWDCheckerType:            model.AWDCheckerTypeLegacyProbe,
		AWDCheckerConfig:          `{}`,
		AWDCheckerValidationState: model.AWDCheckerValidationStateStale,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	})
	createAWDReadinessRelationFixture(t, db, &model.ContestChallenge{
		ContestID:                 701,
		ChallengeID:               7014,
		Points:                    100,
		IsVisible:                 true,
		AWDCheckerType:            model.AWDCheckerTypeHTTPStandard,
		AWDCheckerConfig:          `{"get_flag":{"path":"/health"}}`,
		AWDCheckerValidationState: model.AWDCheckerValidationStatePassed,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	})

	resp, err := service.GetReadiness(context.Background(), 701)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if resp.TotalChallenges != 4 || resp.PassedChallenges != 1 {
		t.Fatalf("unexpected readiness counts: %+v", resp)
	}
	if resp.PendingChallenges != 1 || resp.FailedChallenges != 1 || resp.StaleChallenges != 1 {
		t.Fatalf("unexpected blocking state counts: %+v", resp)
	}
	if resp.BlockingCount != 3 {
		t.Fatalf("expected blocking count 3, got %d", resp.BlockingCount)
	}
	if got := readinessBlockingReasonByChallenge(resp.Items, 7011); got != "last_preview_failed" {
		t.Fatalf("expected last_preview_failed for 7011, got %q", got)
	}
	if got := readinessBlockingReasonByChallenge(resp.Items, 7012); got != "pending_validation" {
		t.Fatalf("expected pending_validation for 7012, got %q", got)
	}
	if got := readinessBlockingReasonByChallenge(resp.Items, 7013); got != "validation_stale" {
		t.Fatalf("expected validation_stale for 7013, got %q", got)
	}
	if len(resp.Items) > 0 && resp.Items[0].LastAccessURL == nil {
		t.Fatalf("expected last_access_url for preview-backed item: %+v", resp.Items[0])
	}
}

func TestAWDQueryServiceGetReadinessTreatsZeroChallengesAsGlobalBlock(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Now()

	createAWDReadinessContestFixture(t, db, 702, now)

	resp, err := service.GetReadiness(context.Background(), 702)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if resp.TotalChallenges != 0 || resp.PassedChallenges != 0 {
		t.Fatalf("unexpected readiness counts: %+v", resp)
	}
	if resp.BlockingCount != 1 {
		t.Fatalf("expected blocking count 1, got %d", resp.BlockingCount)
	}
	if len(resp.GlobalBlockingReasons) != 1 || resp.GlobalBlockingReasons[0] != "no_challenges" {
		t.Fatalf("unexpected global blocking reasons: %+v", resp.GlobalBlockingReasons)
	}
}

func TestAWDQueryServiceGetReadinessTreatsBrokenCheckerConfigAsMissingChecker(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Now()

	createAWDReadinessContestFixture(t, db, 703, now)
	createAWDReadinessChallengeFixture(t, db, 7031, "broken-config", now)
	createAWDReadinessRelationFixture(t, db, &model.ContestChallenge{
		ContestID:                 703,
		ChallengeID:               7031,
		Points:                    100,
		IsVisible:                 true,
		AWDCheckerType:            model.AWDCheckerTypeHTTPStandard,
		AWDCheckerConfig:          `{"get_flag":`,
		AWDCheckerValidationState: model.AWDCheckerValidationStatePassed,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	})

	resp, err := service.GetReadiness(context.Background(), 703)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if resp.MissingCheckerChallenges != 1 {
		t.Fatalf("expected missing checker count 1, got %+v", resp)
	}
	if resp.BlockingCount != 1 {
		t.Fatalf("expected blocking count 1, got %d", resp.BlockingCount)
	}
	if len(resp.Items) != 1 || resp.Items[0].BlockingReason != "invalid_checker_config" {
		t.Fatalf("unexpected readiness items: %+v", resp.Items)
	}
}

func TestAWDQueryServiceGetReadinessItemJSONIncludesRequiredNullableKeys(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Now()

	createAWDReadinessContestFixture(t, db, 704, now)
	createAWDReadinessChallengeFixture(t, db, 7041, "missing-checker", now)
	createAWDReadinessRelationFixture(t, db, &model.ContestChallenge{
		ContestID:                 704,
		ChallengeID:               7041,
		Points:                    100,
		IsVisible:                 true,
		AWDCheckerType:            "",
		AWDCheckerConfig:          `{}`,
		AWDCheckerValidationState: model.AWDCheckerValidationStatePending,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	})

	resp, err := service.GetReadiness(context.Background(), 704)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if len(resp.BlockingActions) != 3 || resp.BlockingActions[1] != "run_current_round_check" {
		t.Fatalf("unexpected blocking actions: %+v", resp.BlockingActions)
	}
	if len(resp.Items) != 1 || resp.Items[0].BlockingReason != "missing_checker" {
		t.Fatalf("unexpected readiness items: %+v", resp.Items)
	}

	raw, err := json.Marshal(resp.Items[0])
	if err != nil {
		t.Fatalf("marshal readiness item: %v", err)
	}

	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal readiness item: %v", err)
	}
	if _, ok := payload["checker_type"]; !ok {
		t.Fatalf("expected checker_type key in payload: %s", string(raw))
	}
	if _, ok := payload["last_preview_at"]; !ok {
		t.Fatalf("expected last_preview_at key in payload: %s", string(raw))
	}
	if _, ok := payload["last_access_url"]; !ok {
		t.Fatalf("expected last_access_url key in payload: %s", string(raw))
	}
}

func createAWDReadinessContestFixture(t *testing.T, db *gorm.DB, contestID int64, now time.Time) {
	t.Helper()

	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "awd-readiness",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
}

func createAWDReadinessChallengeFixture(t *testing.T, db *gorm.DB, challengeID int64, title string, now time.Time) {
	t.Helper()

	if err := db.Create(&model.Challenge{
		ID:         challengeID,
		Title:      title,
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
}

func createAWDReadinessRelationFixture(t *testing.T, db *gorm.DB, relation *model.ContestChallenge) {
	t.Helper()

	if err := db.Create(relation).Error; err != nil {
		t.Fatalf("create contest challenge: %v", err)
	}
}

func readinessBlockingReasonByChallenge(items []*dto.AWDReadinessItemResp, challengeID int64) string {
	for _, item := range items {
		if item.ChallengeID == challengeID {
			return item.BlockingReason
		}
	}
	return ""
}
