package commands

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func newContestAWDServiceForTest(t *testing.T) (*ContestAWDServiceService, *challengeinfra.Repository, *contestinfra.Repository, *contestinfra.ChallengeRepository, *contestinfra.AWDRepository) {
	t.Helper()
	return newContestAWDServiceForTestWithRedis(t, nil)
}

func newContestAWDServiceForTestWithRedis(t *testing.T, redisClient *redis.Client) (*ContestAWDServiceService, *challengeinfra.Repository, *contestinfra.Repository, *contestinfra.ChallengeRepository, *contestinfra.AWDRepository) {
	t.Helper()

	db := contesttestsupport.SetupContestTestDB(t)
	challengeRepo := challengeinfra.NewRepository(db)
	contestRepo := contestinfra.NewRepository(db)
	contestChallengeRepo := contestinfra.NewChallengeRepository(db)
	awdRepo := contestinfra.NewAWDRepository(db)

	return NewContestAWDServiceService(awdRepo, contestRepo, contestChallengeRepo, challengeRepo, challengeRepo, redisClient), challengeRepo, contestRepo, contestChallengeRepo, awdRepo
}

func TestContestAWDServiceServiceCreateFromTemplate(t *testing.T) {
	service, challengeRepo, contestRepo, contestChallengeRepo, awdRepo := newContestAWDServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        801,
		Title:     "awd-service-association",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
		ID:         9801,
		Title:      "bank-portal",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             1001,
		Name:           "Bank Portal",
		Slug:           "bank-portal",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"path":"/internal/flag"}}`,
		AccessConfig:   `{"primary_url":"http://bank.internal"}`,
		RuntimeConfig:  `{"workspace_mode":"per_team"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 801, &dto.CreateContestAWDServiceReq{
		TemplateID:  1001,
		Points:      100,
		DisplayName: "Bank Portal",
		Order:       1,
		IsVisible:   boolPtr(true),
	})
	if err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}
	if resp.TemplateID == nil || *resp.TemplateID != 1001 {
		t.Fatalf("unexpected template id: %+v", resp.TemplateID)
	}
	if resp.ChallengeID != 1001 {
		t.Fatalf("unexpected challenge id: %d", resp.ChallengeID)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), 801, resp.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}
	if stored.TemplateID == nil || *stored.TemplateID != 1001 {
		t.Fatalf("unexpected stored template id: %+v", stored.TemplateID)
	}
	if stored.DisplayName != "Bank Portal" {
		t.Fatalf("unexpected display name: %s", stored.DisplayName)
	}

	if _, err := contestChallengeRepo.FindChallenge(context.Background(), 801, 1001); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected contest_challenges bridge removed, got err=%v", err)
	}
}

func TestContestAWDServiceServiceCreateAppliesDefaultScoreContract(t *testing.T) {
	service, challengeRepo, contestRepo, _, awdRepo := newContestAWDServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        1808,
		Title:     "awd-service-default-score",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             180801,
		Name:           "Default Score Service",
		Slug:           "default-score-service",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"path":"/flag"}}`,
		AccessConfig:   `{"primary_url":"http://default.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 1808, &dto.CreateContestAWDServiceReq{
		TemplateID: 180801,
		Points:     100,
	})
	if err != nil {
		t.Fatalf("CreateContestAWDService() error = %v", err)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), 1808, resp.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}
	var scoreConfig map[string]any
	if err := json.Unmarshal([]byte(stored.ScoreConfig), &scoreConfig); err != nil {
		t.Fatalf("unmarshal score config: %v", err)
	}
	if scoreConfig["awd_sla_score"] != float64(1) || scoreConfig["awd_defense_score"] != float64(2) {
		t.Fatalf("unexpected default score config: %+v", scoreConfig)
	}
}

func TestContestAWDServiceServiceCreateRejectsOversizedServiceScores(t *testing.T) {
	service, challengeRepo, contestRepo, _, _ := newContestAWDServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        1809,
		Title:     "awd-service-oversized-score",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             180901,
		Name:           "Oversized Score Service",
		Slug:           "oversized-score-service",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"path":"/flag"}}`,
		AccessConfig:   `{"primary_url":"http://oversized.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	_, err := service.CreateContestAWDService(context.Background(), 1809, &dto.CreateContestAWDServiceReq{
		TemplateID:      180901,
		Points:          100,
		AWDSLAScore:     intPtr(6),
		AWDDefenseScore: intPtr(2),
	})
	if err == nil {
		t.Fatal("expected oversized SLA score to be rejected")
	}
}

func TestContestAWDServiceServiceCreateRejectsOversizedDisplayPoints(t *testing.T) {
	service, challengeRepo, contestRepo, _, _ := newContestAWDServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        1810,
		Title:     "awd-service-oversized-points",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             181001,
		Name:           "Oversized Points Service",
		Slug:           "oversized-points-service",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"path":"/flag"}}`,
		AccessConfig:   `{"primary_url":"http://points.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	_, err := service.CreateContestAWDService(context.Background(), 1810, &dto.CreateContestAWDServiceReq{
		TemplateID: 181001,
		Points:     501,
	})
	if err == nil {
		t.Fatal("expected oversized display points to be rejected")
	}
}

func TestContestAWDServiceServiceUpdateMaintainsSnapshotOnly(t *testing.T) {
	service, challengeRepo, contestRepo, contestChallengeRepo, awdRepo := newContestAWDServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        802,
		Title:     "awd-service-update-by-id",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
		ID:         9802,
		Title:      "billing-api",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             1002,
		Name:           "Billing API",
		Slug:           "billing-api",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"health":{"path":"/healthz"}}`,
		AccessConfig:   `{"primary_url":"http://billing.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 802, &dto.CreateContestAWDServiceReq{
		TemplateID:  1002,
		Points:      100,
		DisplayName: "Billing API",
		Order:       2,
		IsVisible:   boolPtr(true),
	})
	if err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	visible := false
	displayName := "Billing API v2"
	order := 5
	if err := service.UpdateContestAWDService(context.Background(), 802, resp.ID, &dto.UpdateContestAWDServiceReq{
		DisplayName: &displayName,
		Order:       &order,
		IsVisible:   &visible,
	}); err != nil {
		t.Fatalf("UpdateContestAWDService() error = %v", err)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), 802, resp.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}
	if stored.DisplayName != "Billing API v2" || stored.Order != 5 {
		t.Fatalf("unexpected updated service: %+v", stored)
	}

	if _, err := contestChallengeRepo.FindChallenge(context.Background(), 802, 1002); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected contest_challenges bridge removed, got err=%v", err)
	}
}

func TestContestAWDServiceServiceCreateDoesNotPersistLegacyChallengeIDInRuntimeConfig(t *testing.T) {
	service, challengeRepo, contestRepo, _, awdRepo := newContestAWDServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        804,
		Title:     "awd-service-create-runtime-fields",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
		ID:         9804,
		Title:      "orders-api",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             1004,
		Name:           "Orders API",
		Slug:           "orders-api",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeLegacyProbe,
		CheckerConfig:  `{"health":{"path":"/ready"}}`,
		AccessConfig:   `{"primary_url":"http://orders.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 804, &dto.CreateContestAWDServiceReq{
		TemplateID:      1004,
		Points:          100,
		Order:           1,
		IsVisible:       boolPtr(true),
		CheckerType:     stringPtr(string(model.AWDCheckerTypeHTTPStandard)),
		CheckerConfig:   map[string]any{"get_flag": map[string]any{"path": "/flag"}},
		AWDSLAScore:     intPtr(2),
		AWDDefenseScore: intPtr(3),
	})
	if err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), 804, resp.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}

	var runtimeConfig map[string]any
	if err := json.Unmarshal([]byte(stored.RuntimeConfig), &runtimeConfig); err != nil {
		t.Fatalf("unmarshal runtime config: %v", err)
	}
	if runtimeConfig["checker_type"] != string(model.AWDCheckerTypeHTTPStandard) {
		t.Fatalf("unexpected checker type: %+v", runtimeConfig)
	}
	if _, ok := runtimeConfig["challenge_id"]; ok {
		t.Fatalf("expected runtime config to stop persisting legacy challenge_id, got %+v", runtimeConfig)
	}
	checkerConfig, ok := runtimeConfig["checker_config"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected checker config payload: %+v", runtimeConfig)
	}
	if getFlag, ok := checkerConfig["get_flag"].(map[string]any); !ok || getFlag["path"] != "/flag" {
		t.Fatalf("unexpected checker config: %+v", checkerConfig)
	}

	var scoreConfig map[string]any
	if err := json.Unmarshal([]byte(stored.ScoreConfig), &scoreConfig); err != nil {
		t.Fatalf("unmarshal score config: %v", err)
	}
	if scoreConfig["awd_sla_score"] != float64(2) || scoreConfig["awd_defense_score"] != float64(3) {
		t.Fatalf("unexpected score config: %+v", scoreConfig)
	}
	if stored.ValidationState != model.AWDCheckerValidationStatePending {
		t.Fatalf("unexpected validation state: %s", stored.ValidationState)
	}
}

func TestContestAWDServiceServiceUpdateDoesNotPersistLegacyChallengeIDInRuntimeConfig(t *testing.T) {
	service, challengeRepo, contestRepo, _, awdRepo := newContestAWDServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        805,
		Title:     "awd-service-update-runtime-fields",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
		ID:         9805,
		Title:      "inventory-api",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             1005,
		Name:           "Inventory API",
		Slug:           "inventory-api",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeLegacyProbe,
		CheckerConfig:  `{"health":{"path":"/ready"}}`,
		AccessConfig:   `{"primary_url":"http://inventory.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 805, &dto.CreateContestAWDServiceReq{
		TemplateID: 1005,
		Points:     100,
		Order:      2,
		IsVisible:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	if err := service.UpdateContestAWDService(context.Background(), 805, resp.ID, &dto.UpdateContestAWDServiceReq{
		CheckerType:     stringPtr(string(model.AWDCheckerTypeHTTPStandard)),
		CheckerConfig:   map[string]any{"get_flag": map[string]any{"path": "/healthz"}},
		AWDSLAScore:     intPtr(2),
		AWDDefenseScore: intPtr(4),
	}); err != nil {
		t.Fatalf("UpdateContestAWDService() error = %v", err)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), 805, resp.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}

	var runtimeConfig map[string]any
	if err := json.Unmarshal([]byte(stored.RuntimeConfig), &runtimeConfig); err != nil {
		t.Fatalf("unmarshal runtime config: %v", err)
	}
	if runtimeConfig["checker_type"] != string(model.AWDCheckerTypeHTTPStandard) {
		t.Fatalf("unexpected checker type: %+v", runtimeConfig)
	}
	if _, ok := runtimeConfig["challenge_id"]; ok {
		t.Fatalf("expected runtime config to stop persisting legacy challenge_id, got %+v", runtimeConfig)
	}

	var scoreConfig map[string]any
	if err := json.Unmarshal([]byte(stored.ScoreConfig), &scoreConfig); err != nil {
		t.Fatalf("unmarshal score config: %v", err)
	}
	if scoreConfig["awd_sla_score"] != float64(2) || scoreConfig["awd_defense_score"] != float64(4) {
		t.Fatalf("unexpected score config: %+v", scoreConfig)
	}
}

func TestContestAWDServiceServiceCreateConsumesCheckerPreviewToken(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service, challengeRepo, contestRepo, _, awdRepo := newContestAWDServiceForTestWithRedis(t, redisClient)

	now := time.Now().UTC()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        806,
		Title:     "awd-service-preview-token",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
		ID:         9806,
		Title:      "preview-service",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             1006,
		Name:           "Preview Service",
		Slug:           "preview-service",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"path":"/ready"}}`,
		AccessConfig:   `{"primary_url":"http://preview.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	checkerConfig := map[string]any{
		"get_flag": map[string]any{
			"path": "/flag",
		},
	}
	rawCheckerConfig, err := contestdomain.MarshalAWDCheckerConfig(checkerConfig)
	if err != nil {
		t.Fatalf("marshal checker config: %v", err)
	}
	token, err := storeAWDCheckerPreviewToken(
		context.Background(),
		redisClient,
		806,
		0,
		1006,
		model.AWDCheckerTypeHTTPStandard,
		rawCheckerConfig,
		&dto.AWDCheckerPreviewResp{
			CheckerType:   model.AWDCheckerTypeHTTPStandard,
			ServiceStatus: model.AWDServiceStatusUp,
			CheckResult: map[string]any{
				"checked_at": now.Format(time.RFC3339),
			},
			PreviewContext: dto.AWDCheckerPreviewContextResp{
				AccessURL:   "http://preview.internal",
				PreviewFlag: "flag{preview}",
				ChallengeID: 1006,
			},
		},
	)
	if err != nil {
		t.Fatalf("storeAWDCheckerPreviewToken() error = %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 806, &dto.CreateContestAWDServiceReq{
		TemplateID:             1006,
		Points:                 100,
		Order:                  1,
		IsVisible:              boolPtr(true),
		CheckerType:            stringPtr(string(model.AWDCheckerTypeHTTPStandard)),
		CheckerConfig:          checkerConfig,
		AWDCheckerPreviewToken: stringPtr(token),
	})
	if err != nil {
		t.Fatalf("CreateContestAWDService() error = %v", err)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), 806, resp.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}
	if stored.ValidationState != model.AWDCheckerValidationStatePassed {
		t.Fatalf("unexpected validation state: %s", stored.ValidationState)
	}
	if stored.LastPreviewAt == nil {
		t.Fatal("expected persisted preview time")
	}
	if stored.LastPreviewResult == "" {
		t.Fatal("expected persisted preview result")
	}
}

func TestContestAWDServiceServiceCreateRejectsMissingCheckerPreviewToken(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service, challengeRepo, contestRepo, _, _ := newContestAWDServiceForTestWithRedis(t, redisClient)

	now := time.Now().UTC()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        1806,
		Title:     "awd-service-preview-token-missing",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
		ID:         19806,
		Title:      "preview-service-missing-token",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             1106,
		Name:           "Preview Service Missing Token",
		Slug:           "preview-service-missing-token",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"path":"/ready"}}`,
		AccessConfig:   `{"primary_url":"http://preview-missing.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	_, err = service.CreateContestAWDService(context.Background(), 1806, &dto.CreateContestAWDServiceReq{
		TemplateID:             1106,
		Points:                 100,
		Order:                  1,
		IsVisible:              boolPtr(true),
		CheckerType:            stringPtr(string(model.AWDCheckerTypeHTTPStandard)),
		CheckerConfig:          map[string]any{"get_flag": map[string]any{"path": "/flag"}},
		AWDCheckerPreviewToken: stringPtr("missing-preview-token"),
	})
	if err == nil {
		t.Fatal("expected CreateContestAWDService() to reject missing preview token")
	}
	if !strings.Contains(err.Error(), "试跑结果已失效") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContestAWDServiceServiceUpdateConsumesCheckerPreviewTokenByServiceID(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service, challengeRepo, contestRepo, _, awdRepo := newContestAWDServiceForTestWithRedis(t, redisClient)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        807,
		Title:     "awd-service-update-preview-token-by-service-id",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
		ID:         9807,
		Title:      "preview-update-service",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             1007,
		Name:           "Preview Update Service",
		Slug:           "preview-update-service",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"path":"/ready"}}`,
		AccessConfig:   `{"primary_url":"http://preview-update.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 807, &dto.CreateContestAWDServiceReq{
		TemplateID: 1007,
		Points:     100,
		Order:      1,
		IsVisible:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateContestAWDService() error = %v", err)
	}

	checkerConfig := map[string]any{
		"get_flag": map[string]any{
			"path": "/ready",
		},
	}
	rawCheckerConfig, err := contestdomain.MarshalAWDCheckerConfig(checkerConfig)
	if err != nil {
		t.Fatalf("marshal checker config: %v", err)
	}
	token, err := storeAWDCheckerPreviewToken(
		context.Background(),
		redisClient,
		807,
		resp.ID,
		1007,
		model.AWDCheckerTypeHTTPStandard,
		rawCheckerConfig,
		&dto.AWDCheckerPreviewResp{
			CheckerType:   model.AWDCheckerTypeHTTPStandard,
			ServiceStatus: model.AWDServiceStatusUp,
			CheckResult: map[string]any{
				"checked_at": now.Format(time.RFC3339),
			},
			PreviewContext: dto.AWDCheckerPreviewContextResp{
				ServiceID:   resp.ID,
				AccessURL:   "http://preview-update.internal",
				PreviewFlag: "flag{preview}",
				ChallengeID: 1007,
			},
		},
	)
	if err != nil {
		t.Fatalf("storeAWDCheckerPreviewToken() error = %v", err)
	}

	if err := service.UpdateContestAWDService(context.Background(), 807, resp.ID, &dto.UpdateContestAWDServiceReq{
		AWDCheckerPreviewToken: stringPtr(token),
	}); err != nil {
		t.Fatalf("UpdateContestAWDService() error = %v", err)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), 807, resp.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}
	if stored.ValidationState != model.AWDCheckerValidationStatePassed {
		t.Fatalf("unexpected validation state: %s", stored.ValidationState)
	}
	if stored.LastPreviewAt == nil {
		t.Fatal("expected persisted preview time")
	}
	if stored.LastPreviewResult == "" {
		t.Fatal("expected persisted preview result")
	}
}

func TestContestAWDServiceServiceUpdateRejectsMissingCheckerPreviewToken(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service, challengeRepo, contestRepo, _, _ := newContestAWDServiceForTestWithRedis(t, redisClient)

	now := time.Now().UTC()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        1807,
		Title:     "awd-service-update-missing-preview-token",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
		ID:         19807,
		Title:      "preview-update-missing-token",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             1107,
		Name:           "Preview Update Missing Token",
		Slug:           "preview-update-missing-token",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"path":"/ready"}}`,
		AccessConfig:   `{"primary_url":"http://preview-update-missing.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 1807, &dto.CreateContestAWDServiceReq{
		TemplateID: 1107,
		Points:     100,
		Order:      1,
		IsVisible:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateContestAWDService() error = %v", err)
	}

	err = service.UpdateContestAWDService(context.Background(), 1807, resp.ID, &dto.UpdateContestAWDServiceReq{
		AWDCheckerPreviewToken: stringPtr("missing-preview-token"),
	})
	if err == nil {
		t.Fatal("expected UpdateContestAWDService() to reject missing preview token")
	}
	if !strings.Contains(err.Error(), "试跑结果已失效") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContestAWDServiceServiceDeleteRemovesOnlyServiceRecord(t *testing.T) {
	service, challengeRepo, contestRepo, contestChallengeRepo, awdRepo := newContestAWDServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        803,
		Title:     "awd-service-delete-by-id",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
		ID:         9803,
		Title:      "user-center",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), &model.AWDServiceTemplate{
		ID:             1003,
		Name:           "User Center",
		Slug:           "user-center",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"health":{"path":"/ready"}}`,
		AccessConfig:   `{"primary_url":"http://user.internal"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 803, &dto.CreateContestAWDServiceReq{
		TemplateID:  1003,
		Points:      100,
		DisplayName: "User Center",
		Order:       3,
		IsVisible:   boolPtr(true),
	})
	if err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	if err := service.DeleteContestAWDService(context.Background(), 803, resp.ID); err != nil {
		t.Fatalf("DeleteContestAWDService() error = %v", err)
	}

	_, err = awdRepo.FindContestAWDServiceByContestAndID(context.Background(), 803, resp.ID)
	if err == nil {
		t.Fatal("expected service to be deleted")
	}

	if _, err := contestChallengeRepo.FindChallenge(context.Background(), 803, 1003); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected contest_challenges bridge removed, got err=%v", err)
	}
}
