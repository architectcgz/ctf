package queries

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func newContestAWDServiceQueryServiceForTest(t *testing.T) (*ContestAWDServiceQueryService, *challengeinfra.Repository, *contestinfra.Repository, *contestinfra.AWDRepository) {
	t.Helper()

	db := contesttestsupport.SetupContestTestDB(t)
	return NewContestAWDServiceQueryService(
			contestinfra.NewAWDRepository(db),
			contestinfra.NewRepository(db),
		),
		challengeinfra.NewRepository(db),
		contestinfra.NewRepository(db),
		contestinfra.NewAWDRepository(db)
}

func TestContestAWDServiceQueryServiceListContestAWDServicesIncludesValidationState(t *testing.T) {
	service, challengeRepo, contestRepo, awdRepo := newContestAWDServiceQueryServiceForTest(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        801,
		Title:     "awd-service-query",
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
		Title:      "service-query-challenge",
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

	if err := awdRepo.CreateContestAWDService(context.Background(), &model.ContestAWDService{
		ID:                7101,
		ContestID:         801,
		AWDChallengeID:    9801,
		DisplayName:       "Bank Portal",
		Order:             2,
		IsVisible:         true,
		ScoreConfig:       `{"points":100,"awd_sla_score":1,"awd_defense_score":2}`,
		RuntimeConfig:     `{"challenge_id":9801,"checker_type":"http_standard","checker_config":{"get_flag":{"path":"/health"}}}`,
		ValidationState:   model.AWDCheckerValidationStateFailed,
		LastPreviewAt:     &now,
		LastPreviewResult: `{"service_status":"down","check_result":{"status_code":500},"preview_context":{"access_url":"http://preview.internal"}}`,
		CreatedAt:         now,
		UpdatedAt:         now,
	}); err != nil {
		t.Fatalf("create awd service: %v", err)
	}

	resp, err := service.ListContestAWDServices(context.Background(), 801)
	if err != nil {
		t.Fatalf("ListContestAWDServices() error = %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("unexpected service count: %d", len(resp))
	}
	if resp[0].ValidationState != model.AWDCheckerValidationStateFailed {
		t.Fatalf("expected validation state failed, got %+v", resp[0])
	}
	if resp[0].LastPreviewAt == nil || !resp[0].LastPreviewAt.Equal(now) {
		t.Fatalf("expected last preview at %v, got %+v", now, resp[0].LastPreviewAt)
	}
	if resp[0].LastPreviewResult == nil || resp[0].LastPreviewResult.ServiceStatus != model.AWDServiceStatusDown {
		t.Fatalf("expected preview result in response, got %+v", resp[0].LastPreviewResult)
	}
	if resp[0].LastPreviewResult.PreviewContext.AccessURL != "http://preview.internal" {
		t.Fatalf("unexpected preview access url: %+v", resp[0].LastPreviewResult.PreviewContext)
	}
	if _, ok := resp[0].RuntimeConfig["challenge_id"]; ok {
		t.Fatalf("expected runtime config to hide compatibility challenge_id, got %+v", resp[0].RuntimeConfig)
	}
	if resp[0].RuntimeConfig["checker_type"] != "http_standard" {
		t.Fatalf("expected checker_type to remain visible, got %+v", resp[0].RuntimeConfig)
	}
}
