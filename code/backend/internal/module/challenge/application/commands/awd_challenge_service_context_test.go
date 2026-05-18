package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/model"
	challengeentity "ctf-platform/internal/module/challenge/entity"
)

type awdChallengeCommandContextRepoStub struct {
	createFn   func(ctx context.Context, template *challengeentity.AWDChallenge) error
	findByIDFn func(ctx context.Context, id int64) (*challengeentity.AWDChallenge, error)
	updateFn   func(ctx context.Context, template *challengeentity.AWDChallenge) error
	deleteFn   func(ctx context.Context, id int64) error
}

func (s *awdChallengeCommandContextRepoStub) CreateAWDChallenge(ctx context.Context, template *challengeentity.AWDChallenge) error {
	if s.createFn != nil {
		return s.createFn(ctx, template)
	}
	return nil
}

func (s *awdChallengeCommandContextRepoStub) FindAWDChallengeByID(ctx context.Context, id int64) (*challengeentity.AWDChallenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *awdChallengeCommandContextRepoStub) UpdateAWDChallenge(ctx context.Context, template *challengeentity.AWDChallenge) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, template)
	}
	return nil
}

func (s *awdChallengeCommandContextRepoStub) DeleteAWDChallenge(ctx context.Context, id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

type awdChallengeCommandContextKey string

func TestAWDChallengeServiceCreateChallengePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdChallengeCommandContextKey("create")
	expectedCtxValue := "ctx-create"
	createCalled := false
	repo := &awdChallengeCommandContextRepoStub{
		createFn: func(ctx context.Context, template *challengeentity.AWDChallenge) error {
			createCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected create ctx value %v, got %v", expectedCtxValue, got)
			}
			if template.CreatedBy == nil || *template.CreatedBy != 2001 {
				t.Fatalf("unexpected created_by: %+v", template.CreatedBy)
			}
			return nil
		},
	}
	service := NewAWDChallengeService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.CreateChallenge(ctx, 2001, CreateAWDChallengeInput{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyHard,
		Description:    "desc",
		ServiceType:    string(challengeentity.AWDServiceTypeWebHTTP),
		DeploymentMode: string(challengeentity.AWDDeploymentModeSingleContainer),
	})
	if err != nil {
		t.Fatalf("CreateChallenge() error = %v", err)
	}
	if !createCalled {
		t.Fatal("expected create repository to be called")
	}
	if resp == nil {
		t.Fatal("expected create response")
	}
}

func TestAWDChallengeServiceUpdateChallengePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdChallengeCommandContextKey("update")
	expectedCtxValue := "ctx-update"
	findCalled := false
	updateCalled := false
	repo := &awdChallengeCommandContextRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*challengeentity.AWDChallenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengeentity.AWDChallenge{
				ID:             id,
				Name:           "Legacy",
				Slug:           "legacy",
				Category:       "web",
				Difficulty:     model.ChallengeDifficultyEasy,
				ServiceType:    challengeentity.AWDServiceTypeWebHTTP,
				DeploymentMode: challengeentity.AWDDeploymentModeSingleContainer,
				Status:         challengeentity.AWDChallengeStatusDraft,
			}, nil
		},
		updateFn: func(ctx context.Context, template *challengeentity.AWDChallenge) error {
			updateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update ctx value %v, got %v", expectedCtxValue, got)
			}
			if template.Name != "Bank Portal AWD" || template.Status != challengeentity.AWDChallengeStatusPublished {
				t.Fatalf("unexpected updated template payload: %+v", template)
			}
			return nil
		},
	}
	service := NewAWDChallengeService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UpdateChallenge(ctx, 99, UpdateAWDChallengeInput{
		Name:   "Bank Portal AWD",
		Status: string(challengeentity.AWDChallengeStatusPublished),
	})
	if err != nil {
		t.Fatalf("UpdateChallenge() error = %v", err)
	}
	if !findCalled || !updateCalled {
		t.Fatalf("expected repository calls, got find=%v update=%v", findCalled, updateCalled)
	}
	if resp == nil || resp.Name != "Bank Portal AWD" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAWDChallengeServiceDeleteChallengePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdChallengeCommandContextKey("delete")
	expectedCtxValue := "ctx-delete"
	findCalled := false
	deleteCalled := false
	repo := &awdChallengeCommandContextRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*challengeentity.AWDChallenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengeentity.AWDChallenge{ID: id, Name: "Legacy"}, nil
		},
		deleteFn: func(ctx context.Context, id int64) error {
			deleteCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected delete ctx value %v, got %v", expectedCtxValue, got)
			}
			if id != 99 {
				t.Fatalf("unexpected delete id: %d", id)
			}
			return nil
		},
	}
	service := NewAWDChallengeService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteChallenge(ctx, 99); err != nil {
		t.Fatalf("DeleteChallenge() error = %v", err)
	}
	if !findCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got find=%v delete=%v", findCalled, deleteCalled)
	}
}
