package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type awdServiceTemplateCommandContextRepoStub struct {
	createFn   func(ctx context.Context, template *model.AWDServiceTemplate) error
	findByIDFn func(ctx context.Context, id int64) (*model.AWDServiceTemplate, error)
	updateFn   func(ctx context.Context, template *model.AWDServiceTemplate) error
	deleteFn   func(ctx context.Context, id int64) error
}

func (s *awdServiceTemplateCommandContextRepoStub) CreateAWDServiceTemplate(ctx context.Context, template *model.AWDServiceTemplate) error {
	if s.createFn != nil {
		return s.createFn(ctx, template)
	}
	return nil
}

func (s *awdServiceTemplateCommandContextRepoStub) FindAWDServiceTemplateByID(ctx context.Context, id int64) (*model.AWDServiceTemplate, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *awdServiceTemplateCommandContextRepoStub) UpdateAWDServiceTemplate(ctx context.Context, template *model.AWDServiceTemplate) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, template)
	}
	return nil
}

func (s *awdServiceTemplateCommandContextRepoStub) DeleteAWDServiceTemplate(ctx context.Context, id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

type awdServiceTemplateCommandContextKey string

func TestAWDServiceTemplateServiceCreateTemplatePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdServiceTemplateCommandContextKey("create")
	expectedCtxValue := "ctx-create"
	createCalled := false
	repo := &awdServiceTemplateCommandContextRepoStub{
		createFn: func(ctx context.Context, template *model.AWDServiceTemplate) error {
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
	resp, err := service.CreateTemplate(ctx, 2001, &dto.CreateAWDServiceTemplateReq{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyHard,
		Description:    "desc",
		ServiceType:    string(model.AWDServiceTypeWebHTTP),
		DeploymentMode: string(model.AWDDeploymentModeSingleContainer),
	})
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}
	if !createCalled {
		t.Fatal("expected create repository to be called")
	}
	if resp == nil {
		t.Fatal("expected create response")
	}
}

func TestAWDServiceTemplateServiceUpdateTemplatePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdServiceTemplateCommandContextKey("update")
	expectedCtxValue := "ctx-update"
	findCalled := false
	updateCalled := false
	repo := &awdServiceTemplateCommandContextRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.AWDServiceTemplate, error) {
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
		updateFn: func(ctx context.Context, template *model.AWDServiceTemplate) error {
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
	resp, err := service.UpdateTemplate(ctx, 99, &dto.UpdateAWDServiceTemplateReq{
		Name:   "Bank Portal AWD",
		Status: string(model.AWDServiceTemplateStatusPublished),
	})
	if err != nil {
		t.Fatalf("UpdateTemplate() error = %v", err)
	}
	if !findCalled || !updateCalled {
		t.Fatalf("expected repository calls, got find=%v update=%v", findCalled, updateCalled)
	}
	if resp == nil || resp.Name != "Bank Portal AWD" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAWDServiceTemplateServiceDeleteTemplatePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdServiceTemplateCommandContextKey("delete")
	expectedCtxValue := "ctx-delete"
	findCalled := false
	deleteCalled := false
	repo := &awdServiceTemplateCommandContextRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.AWDServiceTemplate, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.AWDServiceTemplate{ID: id, Name: "Legacy"}, nil
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
	service := NewAWDServiceTemplateService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteTemplate(ctx, 99); err != nil {
		t.Fatalf("DeleteTemplate() error = %v", err)
	}
	if !findCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got find=%v delete=%v", findCalled, deleteCalled)
	}
}
