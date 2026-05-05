package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestAWDChallengeQueryServiceListChallenges(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDChallengeQueryService(repo)

	if err := repo.CreateAWDChallenge(context.Background(), &model.AWDChallenge{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyHard,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDChallengeStatusDraft,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	page, err := service.ListChallenges(context.Background(), ListAWDChallengesInput{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListChallenges() error = %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
	if page.Items[0].Slug != "bank-portal-awd" {
		t.Fatalf("unexpected slug: %s", page.Items[0].Slug)
	}
}

func TestAWDChallengeQueryServiceGetChallengeIncludesInheritedRuntimeFields(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDChallengeQueryService(repo)

	if err := repo.CreateAWDChallenge(context.Background(), &model.AWDChallenge{
		ID:               2401,
		Name:             "Bank Portal AWD",
		Slug:             "bank-portal-awd",
		Category:         "web",
		Difficulty:       model.ChallengeDifficultyHard,
		Description:      "multi-step banking target",
		ServiceType:      model.AWDServiceTypeWebHTTP,
		DeploymentMode:   model.AWDDeploymentModeSingleContainer,
		Status:           model.AWDChallengeStatusPublished,
		CheckerType:      model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:    `{"put_flag":{"path":"/api/flag"},"get_flag":{"path":"/api/flag"}}`,
		FlagMode:         "dynamic_team",
		FlagConfig:       `{"flag_prefix":"awd","rotate_interval_sec":120}`,
		DefenseEntryMode: "http",
		AccessConfig:     `{"public_base_url":"http://bank.internal","service_port":8080}`,
		RuntimeConfig:    `{"image_id":9901,"service_port":8080}`,
		ReadinessStatus:  model.AWDReadinessStatusPassed,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	item, err := service.GetChallenge(context.Background(), 2401)
	if err != nil {
		t.Fatalf("GetChallenge() error = %v", err)
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
