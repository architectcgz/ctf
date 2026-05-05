package commands

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func newContestChallengeCommandService(t *testing.T) (*ChallengeService, *challengeinfra.Repository, *contestinfra.Repository, *contestinfra.ChallengeRepository, *contestinfra.AWDRepository) {
	t.Helper()

	return newContestChallengeCommandServiceWithRedis(t, nil)
}

func newContestChallengeCommandServiceWithRedis(t *testing.T, redisClient *redis.Client) (*ChallengeService, *challengeinfra.Repository, *contestinfra.Repository, *contestinfra.ChallengeRepository, *contestinfra.AWDRepository) {
	t.Helper()

	db := contesttestsupport.SetupContestTestDB(t)
	awdRepo := contestinfra.NewAWDRepository(db)
	return NewChallengeService(
			contestinfra.NewChallengeRepository(db),
			challengeinfra.NewRepository(db),
			contestinfra.NewRepository(db),
			awdRepo,
			redisClient,
		),
		challengeinfra.NewRepository(db),
		contestinfra.NewRepository(db),
		contestinfra.NewChallengeRepository(db),
		awdRepo
}

func TestChallengeServiceAddChallengeToAWDContestDoesNotCreateAWDService(t *testing.T) {
	service, challengeRepo, contestRepo, challengeRelationRepo, awdRepo := newContestChallengeCommandService(t)

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
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
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

	resp, err := service.AddChallengeToContest(context.Background(), contest.ID, AddContestChallengeInput{
		ChallengeID: 9001,
		Points:      120,
		Order:       2,
		IsVisible:   boolPtr(true),
	})
	if err != nil {
		t.Fatalf("AddChallengeToContest() error = %v", err)
	}
	if resp.Points != 120 || resp.Order != 2 || !resp.IsVisible {
		t.Fatalf("unexpected awd challenge response: %+v", resp)
	}

	items, err := challengeRelationRepo.ListChallenges(context.Background(), contest.ID, false)
	if err != nil {
		t.Fatalf("ListChallenges() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected challenge count: %d", len(items))
	}
	if items[0].Points != 120 || items[0].Order != 2 || !items[0].IsVisible {
		t.Fatalf("unexpected stored contest challenge: %+v", items[0])
	}

	services, err := awdRepo.ListContestAWDServicesByContest(context.Background(), contest.ID)
	if err != nil {
		t.Fatalf("ListContestAWDServicesByContest() error = %v", err)
	}
	if len(services) != 0 {
		t.Fatalf("expected generic challenge add to not create contest awd service, got %+v", services)
	}
}

func TestChallengeServiceUpdateChallengeDoesNotCreateAWDService(t *testing.T) {
	service, challengeRepo, contestRepo, challengeRelationRepo, awdRepo := newContestChallengeCommandService(t)

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
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
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

	err := service.UpdateChallenge(context.Background(), contest.ID, 9003, UpdateContestChallengeInput{
		Points:    intPtr(140),
		Order:     intPtr(3),
		IsVisible: boolPtr(false),
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
	if items[0].Points != 140 || items[0].Order != 3 || items[0].IsVisible {
		t.Fatalf("unexpected updated contest challenge: %+v", items[0])
	}

	services, err := awdRepo.ListContestAWDServicesByContest(context.Background(), contest.ID)
	if err != nil {
		t.Fatalf("ListContestAWDServicesByContest() error = %v", err)
	}
	if len(services) != 0 {
		t.Fatalf("expected generic challenge update to not create contest awd service, got %+v", services)
	}
}

func TestChallengeServiceRemoveChallengeFromContestDoesNotDeleteAWDService(t *testing.T) {
	service, challengeRepo, contestRepo, challengeRelationRepo, awdRepo := newContestChallengeCommandService(t)

	now := time.Now()
	contest := &model.Contest{
		ID:        506,
		Title:     "awd-remove",
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
	if err := challengeRepo.Create(context.Background(), &model.Challenge{
		ID:         9006,
		Title:      "awd-remove-challenge",
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
		ChallengeID: 9006,
		Points:      100,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("add challenge: %v", err)
	}
	if err := awdRepo.CreateContestAWDService(context.Background(), &model.ContestAWDService{
		ID:              contesttestsupport.DefaultAWDContestServiceID(contest.ID, 9006),
		ContestID:       contest.ID,
		AWDChallengeID:  9006,
		DisplayName:     "AWD Remove Service",
		Order:           1,
		IsVisible:       true,
		ScoreConfig:     `{"points":100}`,
		RuntimeConfig:   `{"checker_type":"http_standard"}`,
		ValidationState: model.AWDCheckerValidationStatePending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	if err := service.RemoveChallengeFromContest(context.Background(), contest.ID, 9006); err != nil {
		t.Fatalf("RemoveChallengeFromContest() error = %v", err)
	}

	exists, err := challengeRelationRepo.Exists(context.Background(), contest.ID, 9006)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Fatal("expected contest challenge relation removed")
	}
	serviceAssociation, err := awdRepo.FindContestAWDServiceByContestAndID(
		context.Background(),
		contest.ID,
		contesttestsupport.DefaultAWDContestServiceID(contest.ID, 9006),
	)
	if err != nil {
		t.Fatalf("expected explicit awd service to remain after generic challenge removal, got %v", err)
	}
	if serviceAssociation.ID != contesttestsupport.DefaultAWDContestServiceID(contest.ID, 9006) {
		t.Fatalf("unexpected retained awd service: %+v", serviceAssociation)
	}
}

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
