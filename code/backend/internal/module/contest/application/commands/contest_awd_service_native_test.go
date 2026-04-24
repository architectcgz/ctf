package commands

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func TestContestAWDServiceServiceCreateUsesTemplateSnapshotOnly(t *testing.T) {
	service, challengeRepo, contestRepo, contestChallengeRepo, awdRepo := newContestAWDServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        1801,
		Title:     "awd-native-service",
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
		ID:               2801,
		Name:             "Bank Portal",
		Slug:             "bank-portal",
		Category:         "web",
		Difficulty:       model.ChallengeDifficultyMedium,
		Description:      "Bank Portal runtime",
		ServiceType:      model.AWDServiceTypeWebHTTP,
		DeploymentMode:   model.AWDDeploymentModeSingleContainer,
		Status:           model.AWDServiceTemplateStatusPublished,
		CheckerType:      model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:    `{"get_flag":{"path":"/internal/flag"}}`,
		FlagMode:         "dynamic",
		FlagConfig:       `{"flag_type":"dynamic","flag_prefix":"awd"}`,
		DefenseEntryMode: "http",
		AccessConfig:     `{"primary_url":"http://bank.internal"}`,
		RuntimeConfig:    `{"image_id":9901,"instance_sharing":"per_team"}`,
		CreatedAt:        now,
		UpdatedAt:        now,
	}); err != nil {
		t.Fatalf("create template: %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 1801, &dto.CreateContestAWDServiceReq{
		TemplateID: 2801,
		Points:     180,
		Order:      1,
		IsVisible:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateContestAWDService() error = %v", err)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), 1801, resp.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}
	if stored.DisplayName != "Bank Portal" {
		t.Fatalf("expected display name copied from template, got %+v", stored)
	}

	var snapshot map[string]any
	if err := json.Unmarshal([]byte(stored.ServiceSnapshot), &snapshot); err != nil {
		t.Fatalf("unmarshal service snapshot: %v", err)
	}
	if snapshot["name"] != "Bank Portal" || snapshot["category"] != "web" {
		t.Fatalf("unexpected service snapshot: %+v", snapshot)
	}

	var scoreConfig map[string]any
	if err := json.Unmarshal([]byte(stored.ScoreConfig), &scoreConfig); err != nil {
		t.Fatalf("unmarshal score config: %v", err)
	}
	if scoreConfig["points"] != float64(180) {
		t.Fatalf("expected points from request, got %+v", scoreConfig)
	}

	_, err = contestChallengeRepo.FindChallenge(context.Background(), 1801, 2801)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected no contest_challenges bridge, got err=%v", err)
	}
}

func TestContestAWDServiceServiceSnapshotRemainsFrozenAfterTemplateUpdate(t *testing.T) {
	service, challengeRepo, contestRepo, _, awdRepo := newContestAWDServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        1802,
		Title:     "awd-native-freeze",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	template := &model.AWDServiceTemplate{
		ID:             2802,
		Name:           "Billing API",
		Slug:           "billing-api",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyMedium,
		Description:    "Billing runtime",
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"health":{"path":"/health"}}`,
		RuntimeConfig:  `{"image_id":9902}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := challengeRepo.CreateAWDServiceTemplate(context.Background(), template); err != nil {
		t.Fatalf("create template: %v", err)
	}

	resp, err := service.CreateContestAWDService(context.Background(), 1802, &dto.CreateContestAWDServiceReq{
		TemplateID: 2802,
		Points:     120,
		Order:      2,
		IsVisible:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateContestAWDService() error = %v", err)
	}

	template.Name = "Billing API v2"
	template.Category = "misc"
	template.RuntimeConfig = `{"image_id":19902}`
	if err := challengeRepo.UpdateAWDServiceTemplate(context.Background(), template); err != nil {
		t.Fatalf("update template: %v", err)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), 1802, resp.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}

	var snapshot map[string]any
	if err := json.Unmarshal([]byte(stored.ServiceSnapshot), &snapshot); err != nil {
		t.Fatalf("unmarshal service snapshot: %v", err)
	}
	if snapshot["name"] != "Billing API" || snapshot["category"] != "web" {
		t.Fatalf("expected frozen snapshot, got %+v", snapshot)
	}
}
