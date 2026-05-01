package queries

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gorm.io/gorm"

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

type awdReadinessRelationSeed struct {
	relation          *model.ContestChallenge
	checkerType       model.AWDCheckerType
	checkerConfig     string
	slaScore          int
	defenseScore      int
	validationState   model.AWDCheckerValidationState
	lastPreviewAt     *time.Time
	lastPreviewResult string
}

func TestAWDQueryServiceGetReadinessCountsBlockingStates(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Now()

	createAWDReadinessContestFixture(t, db, 701, now)
	createAWDReadinessChallengeFixture(t, db, 7011, "preview-failed", now)
	createAWDReadinessChallengeFixture(t, db, 7012, "pending-service", now)
	createAWDReadinessChallengeFixture(t, db, 7013, "stale-service", now)
	createAWDReadinessChallengeFixture(t, db, 7014, "passed-service", now)
	createAWDReadinessRelationFixture(t, db, awdReadinessRelationSeed{
		relation: &model.ContestChallenge{
			ContestID:   701,
			ChallengeID: 7011,
			Points:      100,
			IsVisible:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		checkerType:       model.AWDCheckerTypeHTTPStandard,
		checkerConfig:     `{"get_flag":{"path":"/health"}}`,
		validationState:   model.AWDCheckerValidationStateFailed,
		lastPreviewAt:     &now,
		lastPreviewResult: `{"service_status":"down","preview_context":{"access_url":"http://preview.internal"}}`,
	})
	createAWDReadinessRelationFixture(t, db, awdReadinessRelationSeed{
		relation: &model.ContestChallenge{
			ContestID:   701,
			ChallengeID: 7012,
			Points:      100,
			IsVisible:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		checkerType:     model.AWDCheckerTypeLegacyProbe,
		checkerConfig:   `{}`,
		validationState: model.AWDCheckerValidationStatePending,
	})
	createAWDReadinessRelationFixture(t, db, awdReadinessRelationSeed{
		relation: &model.ContestChallenge{
			ContestID:   701,
			ChallengeID: 7013,
			Points:      100,
			IsVisible:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		checkerType:     model.AWDCheckerTypeLegacyProbe,
		checkerConfig:   `{}`,
		validationState: model.AWDCheckerValidationStateStale,
	})
	createAWDReadinessRelationFixture(t, db, awdReadinessRelationSeed{
		relation: &model.ContestChallenge{
			ContestID:   701,
			ChallengeID: 7014,
			Points:      100,
			IsVisible:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		checkerType:     model.AWDCheckerTypeHTTPStandard,
		checkerConfig:   `{"get_flag":{"path":"/health"}}`,
		validationState: model.AWDCheckerValidationStatePassed,
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

func TestAWDQueryServiceGetReadinessIgnoresChallengeOnlyContestRelation(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Now()

	createAWDReadinessContestFixture(t, db, 706, now)
	createAWDReadinessChallengeFixture(t, db, 7061, "challenge-only", now)
	if err := db.Create(&model.ContestChallenge{
		ContestID:   706,
		ChallengeID: 7061,
		Points:      100,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create challenge-only contest relation: %v", err)
	}

	resp, err := service.GetReadiness(context.Background(), 706)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if resp.TotalChallenges != 0 || resp.PassedChallenges != 0 {
		t.Fatalf("expected challenge-only relation ignored, got %+v", resp)
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
	createAWDReadinessRelationFixture(t, db, awdReadinessRelationSeed{
		relation: &model.ContestChallenge{
			ContestID:   703,
			ChallengeID: 7031,
			Points:      100,
			IsVisible:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		checkerType:     model.AWDCheckerTypeHTTPStandard,
		checkerConfig:   `{"get_flag":`,
		validationState: model.AWDCheckerValidationStatePassed,
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
	createAWDReadinessRelationFixture(t, db, awdReadinessRelationSeed{
		relation: &model.ContestChallenge{
			ContestID:   704,
			ChallengeID: 7041,
			Points:      100,
			IsVisible:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		checkerType:     "",
		checkerConfig:   `{}`,
		validationState: model.AWDCheckerValidationStatePending,
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

	if resp.Items[0].CheckerType != "" {
		t.Fatalf("expected empty checker type for missing checker item: %+v", resp.Items[0])
	}
	if resp.Items[0].LastPreviewAt != nil {
		t.Fatalf("expected nil last preview for missing checker item: %+v", resp.Items[0])
	}
	if resp.Items[0].LastAccessURL != nil {
		t.Fatalf("expected nil last access url for missing checker item: %+v", resp.Items[0])
	}
}

func TestAWDQueryServiceGetReadinessPrefersContestAWDServiceRuntimeConfig(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Now()

	createAWDReadinessContestFixture(t, db, 705, now)
	createAWDReadinessChallengeFixture(t, db, 7051, "service-defined-missing-checker", now)
	createAWDReadinessRelationFixture(t, db, awdReadinessRelationSeed{
		relation: &model.ContestChallenge{
			ContestID:   705,
			ChallengeID: 7051,
			Points:      100,
			IsVisible:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		checkerType:     "",
		checkerConfig:   `{}`,
		validationState: model.AWDCheckerValidationStatePassed,
	})
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		db,
		705,
		7051,
		"Bank Portal",
		model.AWDCheckerTypeHTTPStandard,
		`{"get_flag":{"path":"/service-health","expected_status":200}}`,
		100,
		18,
		28,
		now,
	)

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
	if resp.Items[0].CheckerType != string(model.AWDCheckerTypeHTTPStandard) {
		t.Fatalf("unexpected checker_type: %+v", resp.Items[0])
	}
	if resp.Items[0].Title != "Bank Portal" {
		t.Fatalf("expected display_name preferred in readiness title, got %+v", resp.Items[0])
	}
}

func TestAWDQueryServiceGetReadinessExposesServiceID(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Now()

	createAWDReadinessContestFixture(t, db, 707, now)
	createAWDReadinessChallengeFixture(t, db, 7071, "service-id-ready", now)
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		db,
		707,
		7071,
		"Readiness Service",
		model.AWDCheckerTypeHTTPStandard,
		`{"get_flag":{"path":"/ready"}}`,
		100,
		12,
		18,
		now,
	)

	resp, err := service.GetReadiness(context.Background(), 707)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 readiness item, got %+v", resp.Items)
	}
	expectedServiceID := contesttestsupport.DefaultAWDContestServiceID(707, 7071)
	if resp.Items[0].ServiceID != expectedServiceID {
		t.Fatalf("expected readiness service_id=%d, got %+v", expectedServiceID, resp.Items[0])
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
		AWDChallengeID: 8011,
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
		AWDChallengeID: 8012,
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
		AWDChallengeID: 8011,
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
		AWDChallengeID: 8012,
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
		AWDChallengeID: 8011,
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
	if resp.Services[0].AWDChallengeID != 8011 || resp.Services[0].InstanceID != 1 || resp.Services[0].AccessURL != "" {
		t.Fatalf("expected first own service to expose instance id without raw access url, got %+v", resp.Services[0])
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
	if targetService["reachable"] != true {
		t.Fatalf("expected target service to expose reachability, got %s", string(raw))
	}
	if _, ok := targetService["access_url"]; ok {
		t.Fatalf("expected target service to hide raw access url, got %s", string(raw))
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

func TestAWDServiceGetUserWorkspaceIncludesQueuedOwnServiceWithoutAccessURL(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Date(2026, 4, 12, 16, 0, 0, 0, time.UTC)

	contesttestsupport.CreateAWDContestFixture(t, db, 806, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 8061, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 806, 8061, now)
	contesttestsupport.CreateAWDTeamFixture(t, db, 8601, 806, "Red", now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 806, 8601, 9601, now)

	serviceID := contesttestsupport.DefaultAWDContestServiceID(806, 8061)
	contestID := int64(806)
	teamID := int64(8601)
	if err := db.Create(&model.Instance{
		ID:          61,
		UserID:      9601,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 8061,
		ServiceID:   &serviceID,
		ShareScope:  model.InstanceSharingPerTeam,
		Status:      model.InstanceStatusPending,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create pending awd workspace instance: %v", err)
	}
	if err := db.Create(&model.AWDServiceOperation{
		ID:            71,
		ContestID:     contestID,
		TeamID:        teamID,
		ServiceID:     serviceID,
		InstanceID:    61,
		OperationType: model.AWDServiceOperationTypeRestart,
		RequestedBy:   model.AWDServiceOperationRequestedByUser,
		Reason:        "user_restart",
		SLABillable:   true,
		Status:        model.AWDServiceOperationStatusProvisioning,
		StartedAt:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create awd service operation: %v", err)
	}

	resp, err := service.GetUserWorkspace(context.Background(), 9601, 806)
	if err != nil {
		t.Fatalf("GetUserWorkspace() error = %v", err)
	}
	item := findAWDWorkspaceServiceByID(resp.Services, serviceID)
	if item == nil {
		t.Fatalf("expected pending own service in workspace, got %+v", resp.Services)
	}
	if item.InstanceID != 61 || item.InstanceStatus != model.InstanceStatusPending {
		t.Fatalf("expected pending instance to remain visible, got %+v", item)
	}
	if item.OperationStatus != model.AWDServiceOperationStatusProvisioning || item.OperationType != model.AWDServiceOperationTypeRestart || item.OperationSLABillable == nil || !*item.OperationSLABillable {
		t.Fatalf("expected latest operation in workspace, got %+v", item)
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
		AWDChallengeID: 8021,
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

func TestAWDServiceGetUserWorkspacePrefersContestServicesAndSeedsMissingDefinitions(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Date(2026, 4, 12, 16, 0, 0, 0, time.UTC)

	contesttestsupport.CreateAWDContestFixture(t, db, 803, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 8031, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 8032, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 8033, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 803, 8031, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 803, 8032, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 803, 8033, now)
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		db,
		803,
		8031,
		"Bank Portal",
		model.AWDCheckerTypeHTTPStandard,
		`{"get_flag":{"path":"/health"}}`,
		100,
		18,
		28,
		now,
	)
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		db,
		803,
		8032,
		"Admin Gateway",
		model.AWDCheckerTypeHTTPStandard,
		`{"get_flag":{"path":"/ready"}}`,
		100,
		18,
		28,
		now,
	)
	if err := db.Where("contest_id = ? AND awd_challenge_id = ?", 803, 8033).
		Delete(&model.ContestAWDService{}).Error; err != nil {
		t.Fatalf("delete generated contest awd service definition: %v", err)
	}

	contesttestsupport.CreateAWDTeamFixture(t, db, 8301, 803, "Red", now)
	contesttestsupport.CreateAWDTeamFixture(t, db, 8302, 803, "Blue", now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 803, 8301, 9301, now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 803, 8302, 9302, now)

	seedAWDWorkspaceInstance(t, db, 6, 9301, 803, 8301, 8031, "http://red-bank.internal", now)
	seedAWDWorkspaceInstance(t, db, 7, 9302, 803, 8302, 8031, "http://blue-bank.internal", now)
	seedAWDWorkspaceInstance(t, db, 8, 9302, 803, 8302, 8033, "http://blue-unmapped.internal", now)

	resp, err := service.GetUserWorkspace(context.Background(), 9301, 803)
	if err != nil {
		t.Fatalf("GetUserWorkspace() error = %v", err)
	}
	if len(resp.Services) != 2 {
		t.Fatalf("expected 2 seeded contest services, got %+v", resp.Services)
	}

	serviceIDs := []int64{resp.Services[0].ServiceID, resp.Services[1].ServiceID}
	expectedBankServiceID := contesttestsupport.DefaultAWDContestServiceID(803, 8031)
	expectedAdminServiceID := contesttestsupport.DefaultAWDContestServiceID(803, 8032)
	if serviceIDs[0] != expectedBankServiceID || serviceIDs[1] != expectedAdminServiceID {
		t.Fatalf("expected contest service ids [%d %d], got %+v", expectedBankServiceID, expectedAdminServiceID, serviceIDs)
	}
	if resp.Services[0].AWDChallengeID != 8031 || resp.Services[0].InstanceID != 6 || resp.Services[0].AccessURL != "" {
		t.Fatalf("expected bank portal service bound by contest service mapping, got %+v", resp.Services[0])
	}
	if resp.Services[1].AWDChallengeID != 8032 || resp.Services[1].AccessURL != "" {
		t.Fatalf("expected admin gateway definition seeded without instance url, got %+v", resp.Services[1])
	}
	if len(resp.Targets) != 1 {
		t.Fatalf("expected 1 target team, got %+v", resp.Targets)
	}
	if len(resp.Targets[0].Services) != 1 {
		t.Fatalf("expected unmapped target instance filtered out, got %+v", resp.Targets[0].Services)
	}
	if resp.Targets[0].Services[0].ServiceID != expectedBankServiceID {
		t.Fatalf("expected target service keyed by contest service id, got %+v", resp.Targets[0].Services[0])
	}
}

func TestAWDServiceGetUserWorkspaceMatchesInstancesByPersistedServiceID(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Date(2026, 4, 12, 16, 30, 0, 0, time.UTC)

	contesttestsupport.CreateAWDContestFixture(t, db, 804, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 8041, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 8042, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 804, 8041, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 804, 8042, now)
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		db,
		804,
		8041,
		"Bank Portal",
		model.AWDCheckerTypeHTTPStandard,
		`{"get_flag":{"path":"/health"}}`,
		100,
		18,
		28,
		now,
	)
	contesttestsupport.CreateAWDTeamFixture(t, db, 8401, 804, "Red", now)
	contesttestsupport.CreateAWDTeamFixture(t, db, 8402, 804, "Blue", now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 804, 8401, 9401, now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 804, 8402, 9402, now)

	if err := ensureAWDWorkspaceInstanceServiceIDColumn(db); err != nil {
		t.Fatalf("ensure instances.service_id column: %v", err)
	}

	serviceID := contesttestsupport.DefaultAWDContestServiceID(804, 8041)
	seedAWDWorkspaceInstance(t, db, 9, 9401, 804, 8401, 8042, "http://red-bank.internal", now)
	seedAWDWorkspaceInstance(t, db, 10, 9402, 804, 8402, 8042, "http://blue-bank.internal", now)
	if err := db.Exec("UPDATE instances SET service_id = ? WHERE id IN (?, ?)", serviceID, 9, 10).Error; err != nil {
		t.Fatalf("persist awd instance service ids: %v", err)
	}

	resp, err := service.GetUserWorkspace(context.Background(), 9401, 804)
	if err != nil {
		t.Fatalf("GetUserWorkspace() error = %v", err)
	}
	if len(resp.Services) != 2 {
		t.Fatalf("expected 2 service definitions, got %+v", resp.Services)
	}
	matchedOwnService := findAWDWorkspaceServiceByID(resp.Services, serviceID)
	if matchedOwnService == nil {
		t.Fatalf("expected contest service %d in own services, got %+v", serviceID, resp.Services)
	}
	if matchedOwnService.InstanceID != 9 || matchedOwnService.AccessURL != "" || matchedOwnService.AWDChallengeID != 8041 {
		t.Fatalf("expected own service matched by persisted service_id, got %+v", matchedOwnService)
	}
	generatedOwnService := findAWDWorkspaceServiceByChallenge(resp.Services, 8042)
	if generatedOwnService == nil {
		t.Fatalf("expected generated service definition for challenge 8042, got %+v", resp.Services)
	}
	if generatedOwnService.AccessURL != "" {
		t.Fatalf("expected generated service definition to keep empty access url, got %+v", generatedOwnService)
	}
	if len(resp.Targets) != 1 || len(resp.Targets[0].Services) != 1 {
		t.Fatalf("expected 1 target service matched by persisted service_id, got %+v", resp.Targets)
	}
	if resp.Targets[0].Services[0].ServiceID != serviceID || !resp.Targets[0].Services[0].Reachable {
		t.Fatalf("expected target service matched by persisted service_id, got %+v", resp.Targets[0].Services[0])
	}
	if resp.Targets[0].Services[0].AWDChallengeID != 8041 {
		t.Fatalf("expected contest service metadata to stay on challenge 8041, got %+v", resp.Targets[0].Services[0])
	}
}

func TestAWDServiceGetUserWorkspaceIgnoresLegacyServiceRowsWithoutServiceID(t *testing.T) {
	service, db := newAWDQueryServiceForTest(t)
	now := time.Date(2026, 4, 12, 17, 0, 0, 0, time.UTC)

	contesttestsupport.CreateAWDContestFixture(t, db, 805, now)
	contesttestsupport.CreateAWDRoundFixture(t, db, 80501, 805, 3, 60, 40, now)
	contesttestsupport.CreateAWDChallengeFixture(t, db, 8051, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, 805, 8051, now)
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		db,
		805,
		8051,
		"Bank Portal",
		model.AWDCheckerTypeHTTPStandard,
		`{"get_flag":{"path":"/health"}}`,
		100,
		18,
		28,
		now,
	)
	contesttestsupport.CreateAWDTeamFixture(t, db, 8501, 805, "Red", now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, 805, 8501, 9501, now)

	serviceID := contesttestsupport.DefaultAWDContestServiceID(805, 8051)
	seedAWDWorkspaceServiceRecord(t, db, &model.AWDTeamService{
		RoundID:        80501,
		TeamID:         8501,
		ServiceID:      serviceID,
		AWDChallengeID: 8051,
		ServiceStatus:  model.AWDServiceStatusUp,
		AttackReceived: 1,
		SLAScore:       30,
		DefenseScore:   20,
		AttackScore:    10,
		UpdatedAt:      now.Add(-time.Minute),
		CreatedAt:      now.Add(-2 * time.Minute),
	})
	seedAWDWorkspaceServiceRecord(t, db, &model.AWDTeamService{
		ID:             8050199,
		RoundID:        80501,
		TeamID:         8501,
		ServiceID:      0,
		AWDChallengeID: 8051,
		ServiceStatus:  model.AWDServiceStatusDown,
		AttackReceived: 9,
		SLAScore:       0,
		DefenseScore:   0,
		AttackScore:    0,
		UpdatedAt:      now,
		CreatedAt:      now,
	})

	resp, err := service.GetUserWorkspace(context.Background(), 9501, 805)
	if err != nil {
		t.Fatalf("GetUserWorkspace() error = %v", err)
	}
	if len(resp.Services) != 1 {
		t.Fatalf("expected only explicit contest service in workspace, got %+v", resp.Services)
	}
	if resp.Services[0].ServiceID != serviceID {
		t.Fatalf("expected workspace service_id=%d, got %+v", serviceID, resp.Services[0])
	}
	if resp.Services[0].ServiceStatus != model.AWDServiceStatusUp {
		t.Fatalf("expected explicit service row to stay authoritative, got %+v", resp.Services[0])
	}
	for _, item := range resp.Services {
		if item.ServiceID == 0 {
			t.Fatalf("expected service row without service_id ignored, got %+v", resp.Services)
		}
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

func createAWDReadinessRelationFixture(t *testing.T, db *gorm.DB, seed awdReadinessRelationSeed) {
	t.Helper()

	if err := db.Create(seed.relation).Error; err != nil {
		t.Fatalf("create contest challenge: %v", err)
	}

	var challenge model.Challenge
	if err := db.Where("id = ?", seed.relation.ChallengeID).First(&challenge).Error; err != nil {
		t.Fatalf("load challenge for awd service fixture: %v", err)
	}
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		db,
		seed.relation.ContestID,
		seed.relation.ChallengeID,
		challenge.Title,
		seed.checkerType,
		seed.checkerConfig,
		seed.relation.Points,
		seed.slaScore,
		seed.defenseScore,
		seed.relation.UpdatedAt,
	)
	contesttestsupport.SyncAWDContestServiceReadinessFixture(
		t,
		db,
		seed.relation.ContestID,
		seed.relation.ChallengeID,
		seed.validationState,
		seed.lastPreviewAt,
		seed.lastPreviewResult,
	)
}

func readinessBlockingReasonByChallenge(items []AWDReadinessItem, challengeID int64) string {
	for _, item := range items {
		if item.AWDChallengeID == challengeID {
			return item.BlockingReason
		}
	}
	return ""
}

func seedAWDWorkspaceInstance(t *testing.T, db *gorm.DB, instanceID, userID, contestID, teamID, challengeID int64, accessURL string, now time.Time) {
	t.Helper()
	serviceID := contesttestsupport.DefaultAWDContestServiceID(contestID, challengeID)

	if err := db.Create(&model.Instance{
		ID:          instanceID,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
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

func ensureAWDWorkspaceInstanceServiceIDColumn(db *gorm.DB) error {
	if db.Migrator().HasColumn(&model.Instance{}, "service_id") {
		return nil
	}
	return db.Exec("ALTER TABLE instances ADD COLUMN service_id integer").Error
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

func findAWDWorkspaceEventByDirection(items []*AWDWorkspaceRecentEventResult, direction string) *AWDWorkspaceRecentEventResult {
	for _, item := range items {
		if item.Direction == direction {
			return item
		}
	}
	return nil
}

func findAWDWorkspaceServiceByID(items []*AWDWorkspaceServiceResult, serviceID int64) *AWDWorkspaceServiceResult {
	for _, item := range items {
		if item.ServiceID == serviceID {
			return item
		}
	}
	return nil
}

func findAWDWorkspaceServiceByChallenge(items []*AWDWorkspaceServiceResult, challengeID int64) *AWDWorkspaceServiceResult {
	for _, item := range items {
		if item.AWDChallengeID == challengeID {
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
