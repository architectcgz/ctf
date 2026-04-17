package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestAWDServiceTemplateQueryServiceListTemplates(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDServiceTemplateQueryService(repo)

	if err := repo.CreateAWDServiceTemplate(&model.AWDServiceTemplate{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyHard,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusDraft,
	}); err != nil {
		t.Fatalf("CreateAWDServiceTemplate() error = %v", err)
	}

	page, err := service.ListTemplates(context.Background(), &dto.AWDServiceTemplateQuery{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
	if page.Items[0].Slug != "bank-portal-awd" {
		t.Fatalf("unexpected slug: %s", page.Items[0].Slug)
	}
}
