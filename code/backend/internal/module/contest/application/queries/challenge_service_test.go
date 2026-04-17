package queries

import (
	"context"
	"reflect"
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

func TestChallengeServiceListAdminChallengesIncludesAWDServiceFields(t *testing.T) {
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
		ContestID:        601,
		ChallengeID:      9101,
		Points:           100,
		IsVisible:        true,
		AWDCheckerType:   model.AWDCheckerTypeHTTPStandard,
		AWDCheckerConfig: `{"get_flag":{"path":"/internal/flag"}}`,
		AWDSLAScore:      12,
		AWDDefenseScore:  22,
		CreatedAt:        now,
		UpdatedAt:        now,
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
	if resp[0].AWDCheckerType != model.AWDCheckerTypeHTTPStandard || resp[0].AWDSLAScore != 12 || resp[0].AWDDefenseScore != 22 {
		t.Fatalf("unexpected challenge response: %+v", resp[0])
	}
	if path := resp[0].AWDCheckerConfig["get_flag"].(map[string]any)["path"]; path != "/internal/flag" {
		t.Fatalf("unexpected checker config: %+v", resp[0].AWDCheckerConfig)
	}
	if resp[0].AWDServiceID == nil || resp[0].AWDTemplateID == nil {
		t.Fatalf("expected awd service metadata, got %+v", resp[0])
	}
	if resp[0].AWDServiceDisplayName != "Bank Portal" {
		t.Fatalf("unexpected awd service display name: %s", resp[0].AWDServiceDisplayName)
	}
}

func TestChallengeServiceListAdminChallengesIncludesCheckerValidationState(t *testing.T) {
	service, challengeRepo, contestRepo, challengeRelationRepo, _ := newContestChallengeQueryService(t)

	now := time.Now()
	if err := contestRepo.Create(context.Background(), &model.Contest{
		ID:        602,
		Title:     "awd-query-validation",
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
		ID:         9102,
		Title:      "awd-query-validation-challenge",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	record := &model.ContestChallenge{
		ContestID:        602,
		ChallengeID:      9102,
		Points:           100,
		IsVisible:        true,
		AWDCheckerType:   model.AWDCheckerTypeHTTPStandard,
		AWDCheckerConfig: `{"get_flag":{"path":"/internal/flag"}}`,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	setChallengeQueryModelField(t, record, "AWDCheckerValidationState", "failed")
	setChallengeQueryModelField(t, record, "AWDCheckerLastPreviewAt", &now)
	setChallengeQueryModelField(t, record, "AWDCheckerLastPreviewResult", `{"service_status":"down","preview_context":{"access_url":"http://preview.internal"}}`)
	if err := challengeRelationRepo.AddChallenge(context.Background(), record); err != nil {
		t.Fatalf("add challenge: %v", err)
	}

	resp, err := service.ListAdminChallenges(context.Background(), 602)
	if err != nil {
		t.Fatalf("ListAdminChallenges() error = %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("unexpected challenge count: %d", len(resp))
	}

	respValue := reflect.ValueOf(resp[0]).Elem()
	if state := respValue.FieldByName("AWDCheckerValidationState"); !state.IsValid() || state.String() != "failed" {
		t.Fatalf("expected failed validation state, got %#v", state)
	}
	if previewAt := respValue.FieldByName("AWDCheckerLastPreviewAt"); !previewAt.IsValid() || previewAt.IsNil() {
		t.Fatal("expected preview time in response")
	}
	if previewResult := respValue.FieldByName("AWDCheckerLastPreviewResult"); !previewResult.IsValid() || previewResult.IsNil() {
		t.Fatal("expected preview result in response")
	}
}

func setChallengeQueryModelField(t *testing.T, target *model.ContestChallenge, field string, value any) {
	t.Helper()

	item := reflect.ValueOf(target).Elem().FieldByName(field)
	if !item.IsValid() {
		t.Fatalf("field %s not found", field)
	}
	if !item.CanSet() {
		t.Fatalf("field %s cannot set", field)
	}

	next := reflect.ValueOf(value)
	if next.Type().AssignableTo(item.Type()) {
		item.Set(next)
		return
	}
	if next.Type().ConvertibleTo(item.Type()) {
		item.Set(next.Convert(item.Type()))
		return
	}
	t.Fatalf("field %s type mismatch: have %s want %s", field, next.Type(), item.Type())
}
