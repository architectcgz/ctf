package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func newContestAWDServiceForTest(t *testing.T) (*ContestAWDServiceService, *challengeinfra.Repository, *contestinfra.Repository, *contestinfra.AWDRepository) {
	t.Helper()

	db := contesttestsupport.SetupContestTestDB(t)
	challengeRepo := challengeinfra.NewRepository(db)
	contestRepo := contestinfra.NewRepository(db)
	awdRepo := contestinfra.NewAWDRepository(db)

	return NewContestAWDServiceService(awdRepo, contestRepo, challengeRepo, challengeRepo), challengeRepo, contestRepo, awdRepo
}

func TestContestAWDServiceServiceCreateFromTemplate(t *testing.T) {
	service, challengeRepo, contestRepo, awdRepo := newContestAWDServiceForTest(t)

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
	if err := challengeRepo.Create(&model.Challenge{
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
	if err := challengeRepo.CreateAWDServiceTemplate(&model.AWDServiceTemplate{
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
		ChallengeID: 9801,
		TemplateID:  1001,
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
	if resp.ChallengeID != 9801 {
		t.Fatalf("unexpected challenge id: %d", resp.ChallengeID)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndChallenge(context.Background(), 801, 9801)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndChallenge() error = %v", err)
	}
	if stored.TemplateID == nil || *stored.TemplateID != 1001 {
		t.Fatalf("unexpected stored template id: %+v", stored.TemplateID)
	}
	if stored.DisplayName != "Bank Portal" {
		t.Fatalf("unexpected display name: %s", stored.DisplayName)
	}
}
