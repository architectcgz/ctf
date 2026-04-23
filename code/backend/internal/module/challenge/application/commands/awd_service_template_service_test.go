package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestAWDServiceTemplateServiceCreateTemplate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDServiceTemplateService(repo)

	resp, err := service.CreateTemplateWithContext(context.Background(), 2001, &dto.CreateAWDServiceTemplateReq{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyHard,
		Description:    "desc",
		ServiceType:    string(model.AWDServiceTypeWebHTTP),
		DeploymentMode: string(model.AWDDeploymentModeSingleContainer),
	})
	if err != nil {
		t.Fatalf("CreateTemplateWithContext() error = %v", err)
	}
	if resp.ID == 0 {
		t.Fatal("expected created template id")
	}
	if resp.CreatedBy == nil || *resp.CreatedBy != 2001 {
		t.Fatalf("unexpected created_by: %+v", resp.CreatedBy)
	}
	if resp.Status != string(model.AWDServiceTemplateStatusDraft) {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
}

func TestAWDServiceTemplateServiceUpdateTemplate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDServiceTemplateService(repo)

	template := &model.AWDServiceTemplate{
		Name:           "Legacy",
		Slug:           "legacy",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyEasy,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusDraft,
	}
	if err := repo.CreateAWDServiceTemplateWithContext(context.Background(), template); err != nil {
		t.Fatalf("CreateAWDServiceTemplateWithContext() error = %v", err)
	}

	resp, err := service.UpdateTemplateWithContext(context.Background(), template.ID, &dto.UpdateAWDServiceTemplateReq{
		Name:   "Bank Portal AWD",
		Status: string(model.AWDServiceTemplateStatusPublished),
	})
	if err != nil {
		t.Fatalf("UpdateTemplateWithContext() error = %v", err)
	}
	if resp.Name != "Bank Portal AWD" {
		t.Fatalf("unexpected name: %s", resp.Name)
	}
	if resp.Status != string(model.AWDServiceTemplateStatusPublished) {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
}
