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

	if err := repo.CreateAWDServiceTemplateWithContext(context.Background(), &model.AWDServiceTemplate{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyHard,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusDraft,
	}); err != nil {
		t.Fatalf("CreateAWDServiceTemplateWithContext() error = %v", err)
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

func TestAWDServiceTemplateQueryServiceGetTemplateIncludesInheritedRuntimeFields(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDServiceTemplateQueryService(repo)

	if err := repo.CreateAWDServiceTemplateWithContext(context.Background(), &model.AWDServiceTemplate{
		ID:               2401,
		Name:             "Bank Portal AWD",
		Slug:             "bank-portal-awd",
		Category:         "web",
		Difficulty:       model.ChallengeDifficultyHard,
		Description:      "multi-step banking target",
		ServiceType:      model.AWDServiceTypeWebHTTP,
		DeploymentMode:   model.AWDDeploymentModeSingleContainer,
		Status:           model.AWDServiceTemplateStatusPublished,
		CheckerType:      model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:    `{"put_flag":{"path":"/api/flag"},"get_flag":{"path":"/api/flag"}}`,
		FlagMode:         "dynamic_team",
		FlagConfig:       `{"flag_prefix":"awd","rotate_interval_sec":120}`,
		DefenseEntryMode: "http",
		AccessConfig:     `{"public_base_url":"http://bank.internal","service_port":8080}`,
		RuntimeConfig:    `{"image_id":9901,"service_port":8080}`,
		ReadinessStatus:  model.AWDReadinessStatusPassed,
	}); err != nil {
		t.Fatalf("CreateAWDServiceTemplateWithContext() error = %v", err)
	}

	item, err := service.GetTemplate(context.Background(), 2401)
	if err != nil {
		t.Fatalf("GetTemplate() error = %v", err)
	}

	if item.CheckerType != string(model.AWDCheckerTypeHTTPStandard) {
		t.Fatalf("unexpected checker_type: %+v", item)
	}
	if item.FlagMode != "dynamic_team" || item.DefenseEntryMode != "http" {
		t.Fatalf("unexpected inherited modes: %+v", item)
	}
	if item.CheckerConfig["put_flag"] == nil {
		t.Fatalf("expected checker_config in template detail, got %+v", item.CheckerConfig)
	}
	if item.FlagConfig["flag_prefix"] != "awd" {
		t.Fatalf("unexpected flag_config: %+v", item.FlagConfig)
	}
	if item.AccessConfig["service_port"] != float64(8080) {
		t.Fatalf("unexpected access_config: %+v", item.AccessConfig)
	}
	if item.RuntimeConfig["image_id"] != float64(9901) {
		t.Fatalf("unexpected runtime_config: %+v", item.RuntimeConfig)
	}
}
