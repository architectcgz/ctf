package commands

import (
	"context"
	"errors"
	"testing"

	"ctf-platform/internal/apperror"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestAWDChallengeServiceCreateChallenge(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDChallengeService(repo)

	resp, err := service.CreateChallenge(context.Background(), 2001, CreateAWDChallengeInput{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     challengeentity.ChallengeDifficultyHard,
		Description:    "desc",
		ServiceType:    string(challengeentity.AWDServiceTypeWebHTTP),
		DeploymentMode: string(challengeentity.AWDDeploymentModeSingleContainer),
	})
	if err != nil {
		t.Fatalf("CreateChallenge() error = %v", err)
	}
	if resp.ID == 0 {
		t.Fatal("expected created template id")
	}
	if resp.CreatedBy == nil || *resp.CreatedBy != 2001 {
		t.Fatalf("unexpected created_by: %+v", resp.CreatedBy)
	}
	if resp.Status != string(challengeentity.AWDChallengeStatusDraft) {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
}

func TestAWDChallengeServiceUpdateChallenge(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := NewAWDChallengeService(repo)

	template := &challengeentity.AWDChallenge{
		Name:           "Legacy",
		Slug:           "legacy",
		Category:       "web",
		Difficulty:     challengeentity.ChallengeDifficultyEasy,
		ServiceType:    challengeentity.AWDServiceTypeWebHTTP,
		DeploymentMode: challengeentity.AWDDeploymentModeSingleContainer,
		Status:         challengeentity.AWDChallengeStatusDraft,
	}
	if err := repo.CreateAWDChallenge(context.Background(), template); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	resp, err := service.UpdateChallenge(context.Background(), template.ID, UpdateAWDChallengeInput{
		Name:   "Bank Portal AWD",
		Status: string(challengeentity.AWDChallengeStatusPublished),
	})
	if err != nil {
		t.Fatalf("UpdateChallenge() error = %v", err)
	}
	if resp.Name != "Bank Portal AWD" {
		t.Fatalf("unexpected name: %s", resp.Name)
	}
	if resp.Status != string(challengeentity.AWDChallengeStatusPublished) {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
}

func TestAWDChallengeServiceTreatsAWDChallengeNotFoundAsNotFound(t *testing.T) {
	t.Parallel()

	service := NewAWDChallengeService(&awdChallengeCommandContextRepoStub{
		findByIDFn: func(context.Context, int64) (*challengeentity.AWDChallenge, error) {
			return nil, challengeports.ErrAWDChallengeNotFound
		},
	})

	_, err := service.UpdateChallenge(context.Background(), 404, UpdateAWDChallengeInput{Name: "missing"})
	if err == nil {
		t.Fatal("expected challenge not found")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrNotFound.Code {
		t.Fatalf("expected apperror.ErrNotFound, got %v", err)
	}
}
