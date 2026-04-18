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

func TestAWDQueryServiceGetReadinessPrefersContestAWDServiceRuntimeConfig(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Now()

	createAWDReadinessContestFixture(t, db, 705, now)
	createAWDReadinessChallengeFixture(t, db, 7051, "legacy-missing-checker", now)
	createAWDReadinessRelationFixture(t, db, &model.ContestChallenge{
		ContestID:                 705,
		ChallengeID:               7051,
		Points:                    100,
		IsVisible:                 true,
		AWDCheckerType:            "",
		AWDCheckerConfig:          `{}`,
		AWDCheckerValidationState: model.AWDCheckerValidationStatePassed,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	})
	if err := contestinfra.NewAWDRepository(db).CreateContestAWDService(context.Background(), &model.ContestAWDService{
		ContestID:   705,
		ChallengeID: 7051,
		DisplayName: "Bank Portal",
		Order:       0,
		IsVisible:   true,
		ScoreConfig: `{"points":100,"awd_sla_score":18,"awd_defense_score":28}`,
		RuntimeConfig: `{
			"challenge_id":7051,
			"checker_type":"http_standard",
			"checker_config":{
				"get_flag":{"path":"/service-health","expected_status":200}
			}
		}`,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	resp, err := service.GetReadiness(context.Background(), 705)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if !resp.Ready || resp.BlockingCount != 0 || resp.PassedChallenges != 1 {
		t.Fatalf("expected readiness to use contest_awd_services runtime config, got %+v", resp)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("unexpected readiness items: %+v", resp.Items)
	}
	if resp.Items[0].CheckerType != model.AWDCheckerTypeHTTPStandard {
		t.Fatalf("unexpected checker_type: %+v", resp.Items[0])
	}
	if resp.Items[0].Title != "Bank Portal" {
		t.Fatalf("expected display_name preferred in readiness title, got %+v", resp.Items[0])
	}
}

func TestAWDServiceGetUserWorkspaceBuildsOwnServicesTargetsAndRecentEvents(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Date(2026, 4, 12, 15, 0, 0, 0, time.UTC)

	contesttestsupport.CreateAWDContestFixture(t, db, 801, now)
	contesttestsupport.CreateAWDRoundFixture(t, db, 80101, 801, 2, 60, 40, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 8011, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 8012, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 801, 8011, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 801, 8012, now)

	contesttestsupport.CreateAWDTeamFixture(t, db, 8101, 801, "Red", now)
	contesttestsupport.CreateAWDTeamFixture(t, db, 8102, 801, "Blue", now)
	contesttestsupport.CreateAWDTeamFixture(t, db, 8103, 801, "Green", now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 801, 8101, 9001, now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 801, 8102, 9002, now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 801, 8103, 9003, now)

	seedAWDWorkspaceInstance(t, db, 1, 9001, 801, 8101, 8011, "http://red-1.internal", now)
	seedAWDWorkspaceInstance(t, db, 2, 9002, 801, 8102, 8011, "http://blue-1.internal", now)
	seedAWDWorkspaceInstance(t, db, 3, 9003, 801, 8103, 8012, "http://green-2.internal", now)

	seedAWDWorkspaceServiceRecord(t, db, &model.AWDTeamService{
		RoundID:        80101,
		TeamID:         8101,
		ServiceID:      contesttestsupport.DefaultAWDContestServiceID(801, 8011),
		ChallengeID:    8011,
		ServiceStatus:  model.AWDServiceStatusUp,
		CheckResult:    `{"status_reason":"healthy"}`,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		AttackReceived: 0,
		SLAScore:       18,
		DefenseScore:   40,
		AttackScore:    0,
		CreatedAt:      now,
		UpdatedAt:      now.Add(2 * time.Minute),
	})
	seedAWDWorkspaceServiceRecord(t, db, &model.AWDTeamService{
		RoundID:        80101,
		TeamID:         8101,
		ServiceID:      contesttestsupport.DefaultAWDContestServiceID(801, 8012),
		ChallengeID:    8012,
		ServiceStatus:  model.AWDServiceStatusCompromised,
		CheckResult:    `{"status_reason":"flag_mismatch"}`,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		AttackReceived: 1,
		SLAScore:       0,
		DefenseScore:   0,
		AttackScore:    0,
		CreatedAt:      now,
		UpdatedAt:      now.Add(3 * time.Minute),
	})
	seedAWDWorkspaceAttackLog(t, db, &model.AWDAttackLog{
		ID:             1,
		RoundID:        80101,
		AttackerTeamID: 8101,
		VictimTeamID:   8102,
		ServiceID:      contesttestsupport.DefaultAWDContestServiceID(801, 8011),
		ChallengeID:    8011,
		AttackType:     model.AWDAttackTypeFlagCapture,
		Source:         model.AWDAttackSourceSubmission,
		IsSuccess:      true,
		ScoreGained:    60,
		CreatedAt:      now.Add(4 * time.Minute),
	})
	seedAWDWorkspaceAttackLog(t, db, &model.AWDAttackLog{
		ID:             2,
		RoundID:        80101,
		AttackerTeamID: 8102,
		VictimTeamID:   8101,
		ServiceID:      contesttestsupport.DefaultAWDContestServiceID(801, 8012),
		ChallengeID:    8012,
		AttackType:     model.AWDAttackTypeFlagCapture,
		Source:         model.AWDAttackSourceSubmission,
		IsSuccess:      false,
		ScoreGained:    0,
		CreatedAt:      now.Add(5 * time.Minute),
	})
	seedAWDWorkspaceAttackLog(t, db, &model.AWDAttackLog{
		ID:             3,
		RoundID:        80101,
		AttackerTeamID: 8102,
		VictimTeamID:   8103,
		ServiceID:      contesttestsupport.DefaultAWDContestServiceID(801, 8011),
		ChallengeID:    8011,
		AttackType:     model.AWDAttackTypeFlagCapture,
		Source:         model.AWDAttackSourceSubmission,
		IsSuccess:      true,
		ScoreGained:    60,
		CreatedAt:      now.Add(6 * time.Minute),
	})

	resp, err := service.GetUserWorkspace(context.Background(), 9001, 801)
	if err != nil {
		t.Fatalf("GetUserWorkspace() error = %v", err)
	}
	if resp.CurrentRound == nil || resp.CurrentRound.RoundNumber != 2 {
		t.Fatalf("expected current round 2, got %+v", resp.CurrentRound)
	}
	if resp.MyTeam == nil || resp.MyTeam.TeamID != 8101 {
		t.Fatalf("expected red team, got %+v", resp.MyTeam)
	}
	if len(resp.Services) != 2 {
		t.Fatalf("expected 2 own services, got %+v", resp.Services)
	}
	if resp.Services[0].ChallengeID != 8011 || resp.Services[0].AccessURL != "http://red-1.internal" {
		t.Fatalf("expected first own service to include red access url, got %+v", resp.Services[0])
	}
	if len(resp.Targets) != 2 {
		t.Fatalf("expected 2 target teams, got %+v", resp.Targets)
	}
	for _, item := range resp.Targets {
		if item.TeamID == 8101 {
			t.Fatalf("expected self team filtered from targets: %+v", resp.Targets)
		}
	}
	if len(resp.RecentEvents) != 2 {
		t.Fatalf("expected 2 related recent events, got %+v", resp.RecentEvents)
	}

	outgoing := findAWDWorkspaceEventByDirection(resp.RecentEvents, "attack_out")
	if outgoing == nil || outgoing.PeerTeamID != 8102 || !outgoing.IsSuccess {
		t.Fatalf("expected outgoing event against blue, got %+v", outgoing)
	}
	incoming := findAWDWorkspaceEventByDirection(resp.RecentEvents, "attack_in")
	if incoming == nil || incoming.PeerTeamID != 8102 || incoming.IsSuccess {
		t.Fatalf("expected incoming failed event from blue, got %+v", incoming)
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal workspace resp: %v", err)
	}
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal workspace resp: %v", err)
	}
	services, ok := payload["services"].([]any)
	if !ok || len(services) != 2 {
		t.Fatalf("unexpected services payload: %s", string(raw))
	}
	firstService, ok := services[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected first service payload: %#v", services[0])
	}
	if firstService["service_id"] != float64(contesttestsupport.DefaultAWDContestServiceID(801, 8011)) {
		t.Fatalf("expected own service to expose service_id, got %s", string(raw))
	}
	targets, ok := payload["targets"].([]any)
	if !ok || len(targets) != 2 {
		t.Fatalf("unexpected targets payload: %s", string(raw))
	}
	blueTarget := findWorkspaceTargetPayload(targets, 8102)
	if blueTarget == nil {
		t.Fatalf("expected blue target in payload: %s", string(raw))
	}
	targetServices, ok := blueTarget["services"].([]any)
	if !ok || len(targetServices) != 1 {
		t.Fatalf("unexpected blue target services payload: %#v", blueTarget)
	}
	targetService, ok := targetServices[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected target service payload: %#v", targetServices[0])
	}
	if targetService["service_id"] != float64(contesttestsupport.DefaultAWDContestServiceID(801, 8011)) {
		t.Fatalf("expected target service to expose service_id, got %s", string(raw))
	}
	events, ok := payload["recent_events"].([]any)
	if !ok || len(events) != 2 {
		t.Fatalf("unexpected recent events payload: %s", string(raw))
	}
	outgoingPayload := findWorkspaceEventPayload(events, "attack_out")
	if outgoingPayload == nil {
		t.Fatalf("expected outgoing event payload: %s", string(raw))
	}
	if outgoingPayload["service_id"] != float64(contesttestsupport.DefaultAWDContestServiceID(801, 8011)) {
		t.Fatalf("expected outgoing event to expose service_id, got %s", string(raw))
	}
}

func TestAWDServiceGetUserWorkspaceWithoutTeamHidesTargets(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Date(2026, 4, 12, 15, 30, 0, 0, time.UTC)

	contesttestsupport.CreateAWDContestFixture(t, db, 802, now)
	contesttestsupport.CreateAWDRoundFixture(t, db, 80201, 802, 1, 50, 50, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 8021, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 802, 8021, now)
	contesttestsupport.CreateAWDTeamFixture(t, db, 8201, 802, "Alpha", now)
	contesttestsupport.CreateAWDTeamFixture(t, db, 8202, 802, "Beta", now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 802, 8201, 9201, now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 802, 8202, 9202, now)
	seedAWDWorkspaceInstance(t, db, 4, 9201, 802, 8201, 8021, "http://alpha.internal", now)
	seedAWDWorkspaceInstance(t, db, 5, 9202, 802, 8202, 8021, "http://beta.internal", now)
	seedAWDWorkspaceAttackLog(t, db, &model.AWDAttackLog{
		ID:             4,
		RoundID:        80201,
		AttackerTeamID: 8201,
		VictimTeamID:   8202,
		ServiceID:      contesttestsupport.DefaultAWDContestServiceID(802, 8021),
		ChallengeID:    8021,
		AttackType:     model.AWDAttackTypeFlagCapture,
		Source:         model.AWDAttackSourceSubmission,
		IsSuccess:      true,
		ScoreGained:    50,
		CreatedAt:      now.Add(time.Minute),
	})

	resp, err := service.GetUserWorkspace(context.Background(), 9999, 802)
	if err != nil {
		t.Fatalf("GetUserWorkspace() error = %v", err)
	}
	if resp.CurrentRound == nil || resp.CurrentRound.RoundNumber != 1 {
		t.Fatalf("expected current round 1, got %+v", resp.CurrentRound)
	}
	if resp.MyTeam != nil {
		t.Fatalf("expected no team for outsider, got %+v", resp.MyTeam)
	}
	if len(resp.Services) != 0 {
		t.Fatalf("expected no own services for outsider, got %+v", resp.Services)
	}
	if len(resp.Targets) != 0 {
		t.Fatalf("expected no targets for outsider, got %+v", resp.Targets)
	}
	if len(resp.RecentEvents) != 0 {
		t.Fatalf("expected no events for outsider, got %+v", resp.RecentEvents)
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

func seedAWDWorkspaceInstance(t *testing.T, db *gorm.DB, instanceID, userID, contestID, teamID, challengeID int64, accessURL string, now time.Time) {
	t.Helper()

	if err := db.Create(&model.Instance{
		ID:          instanceID,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ContainerID: "container",
		ShareScope:  model.InstanceSharingPerTeam,
		Status:      model.InstanceStatusRunning,
		AccessURL:   accessURL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd workspace instance: %v", err)
	}
}

func seedAWDWorkspaceServiceRecord(t *testing.T, db *gorm.DB, record *model.AWDTeamService) {
	t.Helper()

	if err := db.Create(record).Error; err != nil {
		t.Fatalf("create awd workspace service record: %v", err)
	}
}

func seedAWDWorkspaceAttackLog(t *testing.T, db *gorm.DB, record *model.AWDAttackLog) {
	t.Helper()

	if err := db.Create(record).Error; err != nil {
		t.Fatalf("create awd workspace attack log: %v", err)
	}
}

func findAWDWorkspaceEventByDirection(items []*dto.ContestAWDWorkspaceRecentEventResp, direction string) *dto.ContestAWDWorkspaceRecentEventResp {
	for _, item := range items {
		if item.Direction == direction {
			return item
		}
	}
	return nil
}

func findWorkspaceTargetPayload(items []any, teamID int64) map[string]any {
	for _, item := range items {
		record, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if record["team_id"] == float64(teamID) {
			return record
		}
	}
	return nil
}

func findWorkspaceEventPayload(items []any, direction string) map[string]any {
	for _, item := range items {
		record, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if record["direction"] == direction {
			return record
		}
	}
	return nil
}
