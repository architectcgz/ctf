package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestAWDChallengeServiceCreateChallenge(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDChallengeService(repo)

	resp, err := service.CreateChallenge(context.Background(), 2001, &dto.CreateAWDChallengeReq{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyHard,
		Description:    "desc",
		ServiceType:    string(model.AWDServiceTypeWebHTTP),
		DeploymentMode: string(model.AWDDeploymentModeSingleContainer),
	})
	if err != nil {
		t.Fatalf("CreateChallenge() error = %v", err)
	}
	if resp.ID == 0 {
		t.Fatal("expected created template id")
	}
	if resp.CreatedBy == nil || *resp.CreatedBy != 2001 {
		t.Fatalf("unexpected created_by: %+v", resp.CreatedBy)
	}
	if resp.Status != string(model.AWDChallengeStatusDraft) {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
}

func TestAWDChallengeServiceUpdateChallenge(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDChallengeService(repo)

	template := &model.AWDChallenge{
		Name:           "Legacy",
		Slug:           "legacy",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyEasy,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDChallengeStatusDraft,
	}
	if err := repo.CreateAWDChallenge(context.Background(), template); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	resp, err := service.UpdateChallenge(context.Background(), template.ID, &dto.UpdateAWDChallengeReq{
		Name:   "Bank Portal AWD",
		Status: string(model.AWDChallengeStatusPublished),
	})
	if err != nil {
		t.Fatalf("UpdateChallenge() error = %v", err)
	}
	if resp.Name != "Bank Portal AWD" {
		t.Fatalf("unexpected name: %s", resp.Name)
	}
	if resp.Status != string(model.AWDChallengeStatusPublished) {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
}
