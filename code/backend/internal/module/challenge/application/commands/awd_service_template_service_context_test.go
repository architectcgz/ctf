package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type awdServiceTemplateCommandContextRepoStub struct {
	createWithContextFn   func(ctx context.Context, template *model.AWDServiceTemplate) error
	findByIDWithContextFn func(ctx context.Context, id int64) (*model.AWDServiceTemplate, error)
	updateWithContextFn   func(ctx context.Context, template *model.AWDServiceTemplate) error
	deleteWithContextFn   func(ctx context.Context, id int64) error
}

func (s *awdServiceTemplateCommandContextRepoStub) CreateAWDServiceTemplateWithContext(ctx context.Context, template *model.AWDServiceTemplate) error {
	if s.createWithContextFn != nil {
		return s.createWithContextFn(ctx, template)
	}
	return nil
}

func (s *awdServiceTemplateCommandContextRepoStub) FindAWDServiceTemplateByIDWithContext(ctx context.Context, id int64) (*model.AWDServiceTemplate, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *awdServiceTemplateCommandContextRepoStub) UpdateAWDServiceTemplateWithContext(ctx context.Context, template *model.AWDServiceTemplate) error {
	if s.updateWithContextFn != nil {
		return s.updateWithContextFn(ctx, template)
	}
	return nil
}

func (s *awdServiceTemplateCommandContextRepoStub) DeleteAWDServiceTemplateWithContext(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return nil
}

type awdServiceTemplateCommandContextKey string

func TestAWDServiceTemplateServiceCreateTemplateWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdServiceTemplateCommandContextKey("create")
	expectedCtxValue := "ctx-create"
	createCalled := false
	repo := &awdServiceTemplateCommandContextRepoStub{
		createWithContextFn: func(ctx context.Context, template *model.AWDServiceTemplate) error {
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
	service := NewAWDServiceTemplateService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.CreateTemplateWithContext(ctx, 2001, &dto.CreateAWDServiceTemplateReq{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyHard,
		Description:    "desc",
		ServiceType:    string(model.AWDServiceTypeWebHTTP),
		DeploymentMode: string(model.AWDDeploymentModeSingleContainer),
	})
	if err != nil {
		t.Fatalf("CreateTemplateWithContext() error = %v", err)
	}
	if !createCalled {
		t.Fatal("expected create repository to be called")
	}
	if resp == nil {
		t.Fatal("expected create response")
	}
}

func TestAWDServiceTemplateServiceUpdateTemplateWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdServiceTemplateCommandContextKey("update")
	expectedCtxValue := "ctx-update"
	findCalled := false
	updateCalled := false
	repo := &awdServiceTemplateCommandContextRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.AWDServiceTemplate, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.AWDServiceTemplate{
				ID:             id,
				Name:           "Legacy",
				Slug:           "legacy",
				Category:       "web",
				Difficulty:     model.ChallengeDifficultyEasy,
				ServiceType:    model.AWDServiceTypeWebHTTP,
				DeploymentMode: model.AWDDeploymentModeSingleContainer,
				Status:         model.AWDServiceTemplateStatusDraft,
			}, nil
		},
		updateWithContextFn: func(ctx context.Context, template *model.AWDServiceTemplate) error {
			updateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update ctx value %v, got %v", expectedCtxValue, got)
			}
			if template.Name != "Bank Portal AWD" || template.Status != model.AWDServiceTemplateStatusPublished {
				t.Fatalf("unexpected updated template payload: %+v", template)
			}
			return nil
		},
	}
	service := NewAWDServiceTemplateService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UpdateTemplateWithContext(ctx, 99, &dto.UpdateAWDServiceTemplateReq{
		Name:   "Bank Portal AWD",
		Status: string(model.AWDServiceTemplateStatusPublished),
	})
	if err != nil {
		t.Fatalf("UpdateTemplateWithContext() error = %v", err)
	}
	if !findCalled || !updateCalled {
		t.Fatalf("expected repository calls, got find=%v update=%v", findCalled, updateCalled)
	}
	if resp == nil || resp.Name != "Bank Portal AWD" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAWDServiceTemplateServiceDeleteTemplateWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdServiceTemplateCommandContextKey("delete")
	expectedCtxValue := "ctx-delete"
	findCalled := false
	deleteCalled := false
	repo := &awdServiceTemplateCommandContextRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.AWDServiceTemplate, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.AWDServiceTemplate{ID: id, Name: "Legacy"}, nil
		},
		deleteWithContextFn: func(ctx context.Context, id int64) error {
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
	service := NewAWDServiceTemplateService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteTemplateWithContext(ctx, 99); err != nil {
		t.Fatalf("DeleteTemplateWithContext() error = %v", err)
	}
	if !findCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got find=%v delete=%v", findCalled, deleteCalled)
	}
}
