package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type awdChallengeCommandContextRepoStub struct {
	createFn   func(ctx context.Context, template *model.AWDChallenge) error
	findByIDFn func(ctx context.Context, id int64) (*model.AWDChallenge, error)
	updateFn   func(ctx context.Context, template *model.AWDChallenge) error
	deleteFn   func(ctx context.Context, id int64) error
}

func (s *awdChallengeCommandContextRepoStub) CreateAWDChallenge(ctx context.Context, template *model.AWDChallenge) error {
	if s.createFn != nil {
		return s.createFn(ctx, template)
	}
	return nil
}

func (s *awdChallengeCommandContextRepoStub) FindAWDChallengeByID(ctx context.Context, id int64) (*model.AWDChallenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *awdChallengeCommandContextRepoStub) UpdateAWDChallenge(ctx context.Context, template *model.AWDChallenge) error {
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
		createFn: func(ctx context.Context, template *model.AWDChallenge) error {
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
	resp, err := service.CreateChallenge(ctx, 2001, &dto.CreateAWDChallengeReq{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyHard,
		Description:    "desc",
		ServiceType:    string(model.AWDServiceTypeWebHTTP),
		DeploymentMode: string(model.AWDDeploymentModeSingleContainer),
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
		findByIDFn: func(ctx context.Context, id int64) (*model.AWDChallenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.AWDChallenge{
				ID:             id,
				Name:           "Legacy",
				Slug:           "legacy",
				Category:       "web",
				Difficulty:     model.ChallengeDifficultyEasy,
				ServiceType:    model.AWDServiceTypeWebHTTP,
				DeploymentMode: model.AWDDeploymentModeSingleContainer,
				Status:         model.AWDChallengeStatusDraft,
			}, nil
		},
		updateFn: func(ctx context.Context, template *model.AWDChallenge) error {
			updateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update ctx value %v, got %v", expectedCtxValue, got)
			}
			if template.Name != "Bank Portal AWD" || template.Status != model.AWDChallengeStatusPublished {
				t.Fatalf("unexpected updated template payload: %+v", template)
			}
			return nil
		},
	}
	service := NewAWDChallengeService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UpdateChallenge(ctx, 99, &dto.UpdateAWDChallengeReq{
		Name:   "Bank Portal AWD",
		Status: string(model.AWDChallengeStatusPublished),
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
		findByIDFn: func(ctx context.Context, id int64) (*model.AWDChallenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.AWDChallenge{ID: id, Name: "Legacy"}, nil
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
