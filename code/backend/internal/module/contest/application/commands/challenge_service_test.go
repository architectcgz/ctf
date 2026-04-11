package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	"ctf-platform/pkg/errcode"
)

func newContestChallengeCommandService(t *testing.T) (*ChallengeService, *challengeinfra.Repository, *contestinfra.Repository, *contestinfra.ChallengeRepository) {
	t.Helper()

	db := contesttestsupport.SetupContestTestDB(t)
	return NewChallengeService(
			contestinfra.NewChallengeRepository(db),
			challengeinfra.NewRepository(db),
			contestinfra.NewRepository(db),
		),
		challengeinfra.NewRepository(db),
		contestinfra.NewRepository(db),
		contestinfra.NewChallengeRepository(db)
}

func TestChallengeServiceAddChallengeToAWDContestPersistsAWDServiceConfig(t *testing.T) {
	service, challengeRepo, contestRepo, challengeRelationRepo := newContestChallengeCommandService(t)

	now := time.Now()
	contest := &model.Contest{
		ID:        501,
		Title:     "awd-config",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := contestRepo.Create(context.Background(), contest); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(&model.Challenge{
		ID:         9001,
		Title:      "awd-web",
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

	resp, err := service.AddChallengeToContest(context.Background(), contest.ID, &dto.AddContestChallengeReq{
		ChallengeID:      9001,
		Points:           120,
		AWDCheckerType:   model.AWDCheckerTypeHTTPStandard,
		AWDCheckerConfig: map[string]any{"get_flag": map[string]any{"path": "/internal/flag"}},
		AWDSLAScore:      15,
		AWDDefenseScore:  25,
	})
	if err != nil {
		t.Fatalf("AddChallengeToContest() error = %v", err)
	}
	if resp.AWDCheckerType != model.AWDCheckerTypeHTTPStandard || resp.AWDSLAScore != 15 || resp.AWDDefenseScore != 25 {
		t.Fatalf("unexpected awd challenge response: %+v", resp)
	}
	if path := resp.AWDCheckerConfig["get_flag"].(map[string]any)["path"]; path != "/internal/flag" {
		t.Fatalf("unexpected checker config: %+v", resp.AWDCheckerConfig)
	}

	items, err := challengeRelationRepo.ListChallenges(context.Background(), contest.ID, false)
	if err != nil {
		t.Fatalf("ListChallenges() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected challenge count: %d", len(items))
	}
	if items[0].AWDCheckerType != model.AWDCheckerTypeHTTPStandard || items[0].AWDSLAScore != 15 || items[0].AWDDefenseScore != 25 {
		t.Fatalf("unexpected stored contest challenge: %+v", items[0])
	}
}

func TestChallengeServiceRejectsAWDServiceConfigOnNonAWDContest(t *testing.T) {
	service, challengeRepo, contestRepo, _ := newContestChallengeCommandService(t)

	now := time.Now()
	contest := &model.Contest{
		ID:        502,
		Title:     "jeopardy",
		Mode:      model.ContestModeJeopardy,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := contestRepo.Create(context.Background(), contest); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(&model.Challenge{
		ID:         9002,
		Title:      "web",
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

	_, err := service.AddChallengeToContest(context.Background(), contest.ID, &dto.AddContestChallengeReq{
		ChallengeID:     9002,
		AWDCheckerType:  model.AWDCheckerTypeHTTPStandard,
		AWDSLAScore:     10,
		AWDDefenseScore: 20,
	})
	if err == nil {
		t.Fatal("expected AddChallengeToContest() to reject awd config on non-awd contest")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestChallengeServiceUpdateChallengePersistsAWDServiceConfig(t *testing.T) {
	service, challengeRepo, contestRepo, challengeRelationRepo := newContestChallengeCommandService(t)

	now := time.Now()
	contest := &model.Contest{
		ID:        503,
		Title:     "awd-update",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := contestRepo.Create(context.Background(), contest); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(&model.Challenge{
		ID:         9003,
		Title:      "awd-update-challenge",
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
		ContestID:   contest.ID,
		ChallengeID: 9003,
		Points:      100,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("add challenge: %v", err)
	}

	err := service.UpdateChallenge(context.Background(), contest.ID, 9003, &dto.UpdateContestChallengeReq{
		AWDCheckerType:   stringPtr(string(model.AWDCheckerTypeHTTPStandard)),
		AWDCheckerConfig: map[string]any{"havoc": map[string]any{"path": "/healthz"}},
		AWDSLAScore:      intPtr(18),
		AWDDefenseScore:  intPtr(28),
	})
	if err != nil {
		t.Fatalf("UpdateChallenge() error = %v", err)
	}

	items, err := challengeRelationRepo.ListChallenges(context.Background(), contest.ID, false)
	if err != nil {
		t.Fatalf("ListChallenges() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected challenge count: %d", len(items))
	}
	if items[0].AWDCheckerType != model.AWDCheckerTypeHTTPStandard || items[0].AWDSLAScore != 18 || items[0].AWDDefenseScore != 28 {
		t.Fatalf("unexpected updated contest challenge: %+v", items[0])
	}
}

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
