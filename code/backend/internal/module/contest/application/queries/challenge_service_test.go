package queries

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func newContestChallengeQueryService(t *testing.T) (*ChallengeService, *challengeinfra.Repository, *contestinfra.Repository, *contestinfra.ChallengeRepository, *contestinfra.AWDRepository) {
	t.Helper()

	db := contesttestsupport.SetupContestTestDB(t)
	awdRepo := contestinfra.NewAWDRepository(db)
	return NewChallengeService(
			contestinfra.NewChallengeRepository(db),
			challengeinfra.NewRepository(db),
			contestinfra.NewRepository(db),
			awdRepo,
		),
		challengeinfra.NewRepository(db),
		contestinfra.NewRepository(db),
		contestinfra.NewChallengeRepository(db),
		awdRepo
}

func TestChallengeServiceListAdminChallengesReturnsRelationFieldsOnly(t *testing.T) {
	service, challengeRepo, contestRepo, challengeRelationRepo, awdRepo := newContestChallengeQueryService(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        601,
		Title:     "awd-query",
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
		ID:         9101,
		Title:      "awd-query-challenge",
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
	if err := challengeRelationRepo.AddChallenge(context.Background(), &model.ContestChallenge{
		ContestID:   601,
		ChallengeID: 9101,
		Points:      100,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("add challenge: %v", err)
	}
	templateID := int64(3001)
	if err := awdRepo.CreateContestAWDService(context.Background(), &model.ContestAWDService{
		ContestID:     601,
		ChallengeID:   9101,
		TemplateID:    &templateID,
		DisplayName:   "Bank Portal",
		Order:         0,
		IsVisible:     true,
		ScoreConfig:   `{"points":100,"awd_sla_score":12,"awd_defense_score":22}`,
		RuntimeConfig: `{"challenge_id":9101}`,
		CreatedAt:     now,
		UpdatedAt:     now,
	}); err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	resp, err := service.ListAdminChallenges(context.Background(), 601)
	if err != nil {
		t.Fatalf("ListAdminChallenges() error = %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("unexpected challenge count: %d", len(resp))
	}
	if resp[0].ChallengeID != 9101 || resp[0].Points != 100 || !resp[0].IsVisible {
		t.Fatalf("unexpected challenge response: %+v", resp[0])
	}
	raw, err := json.Marshal(resp[0])
	if err != nil {
		t.Fatalf("marshal challenge response: %v", err)
	}
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal challenge response: %v", err)
	}
	for _, key := range []string{
		"awd_service_id",
		"awd_template_id",
		"awd_service_display_name",
		"awd_checker_type",
		"awd_checker_config",
		"awd_sla_score",
		"awd_defense_score",
		"awd_checker_validation_state",
		"awd_checker_last_preview_at",
		"awd_checker_last_preview_result",
	} {
		if _, ok := payload[key]; ok {
			t.Fatalf("expected admin challenge response without %s, got %s", key, string(raw))
		}
	}
}

func TestChallengeServiceGetContestChallengesReadsAWDServicesFromServiceSnapshot(t *testing.T) {
	service, _, contestRepo, _, awdRepo := newContestChallengeQueryService(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        611,
		Title:     "awd-visible-query",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := awdRepo.CreateContestAWDService(context.Background(), &model.ContestAWDService{
		ID:              7201,
		ContestID:       611,
		ChallengeID:     9111,
		DisplayName:     "Bank Portal",
		Order:           0,
		IsVisible:       true,
		ScoreConfig:     `{"points":100}`,
		ServiceSnapshot: `{"name":"Bank Portal","category":"web","difficulty":"medium","flag_config":{"flag_type":"static","flag_prefix":"awd"}}`,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	resp, err := service.GetContestChallenges(context.Background(), 3001, 611)
	if err != nil {
		t.Fatalf("GetContestChallenges() error = %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("unexpected challenge count: %d", len(resp))
	}
	if resp[0].AWDServiceID == nil || *resp[0].AWDServiceID != 7201 {
		t.Fatalf("expected awd service id 7201, got %+v", resp[0])
	}
	if resp[0].ChallengeID != 9111 || resp[0].Title != "Bank Portal" || resp[0].Category != "web" || resp[0].Difficulty != model.ChallengeDifficultyMedium {
		t.Fatalf("expected awd challenge info from service snapshot, got %+v", resp[0])
	}
}
