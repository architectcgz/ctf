package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type awdServiceTemplateQueryContextRepoStub struct {
	findByIDFn            func(id int64) (*model.AWDServiceTemplate, error)
	findByIDWithContextFn func(ctx context.Context, id int64) (*model.AWDServiceTemplate, error)
	listFn                func(query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error)
	listWithContextFn     func(ctx context.Context, query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error)
}

func (s *awdServiceTemplateQueryContextRepoStub) FindAWDServiceTemplateByID(id int64) (*model.AWDServiceTemplate, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(id)
	}
	return nil, nil
}

func (s *awdServiceTemplateQueryContextRepoStub) FindAWDServiceTemplateByIDWithContext(ctx context.Context, id int64) (*model.AWDServiceTemplate, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindAWDServiceTemplateByID(id)
}

func (s *awdServiceTemplateQueryContextRepoStub) ListAWDServiceTemplates(query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error) {
	if s.listFn != nil {
		return s.listFn(query)
	}
	return nil, 0, nil
}

func (s *awdServiceTemplateQueryContextRepoStub) ListAWDServiceTemplatesWithContext(ctx context.Context, query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error) {
	if s.listWithContextFn != nil {
		return s.listWithContextFn(ctx, query)
	}
	return s.ListAWDServiceTemplates(query)
}

type awdServiceTemplateQueryContextKey string

func TestAWDServiceTemplateQueryServiceGetTemplateWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdServiceTemplateQueryContextKey("get")
	expectedCtxValue := "ctx-get"
	findCalled := false
	repo := &awdServiceTemplateQueryContextRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.AWDServiceTemplate, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.AWDServiceTemplate{ID: id, Name: "Bank Portal AWD", Slug: "bank-portal-awd"}, nil
		},
	}
	service := NewAWDServiceTemplateQueryService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetTemplateWithContext(ctx, 77)
	if err != nil {
		t.Fatalf("GetTemplateWithContext() error = %v", err)
	}
	if !findCalled {
		t.Fatal("expected find repository to be called")
	}
	if resp == nil || resp.ID != 77 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAWDServiceTemplateQueryServiceListTemplatesWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdServiceTemplateQueryContextKey("list")
	expectedCtxValue := "ctx-list"
	listCalled := false
	repo := &awdServiceTemplateQueryContextRepoStub{
		listWithContextFn: func(ctx context.Context, query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error) {
			listCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list ctx value %v, got %v", expectedCtxValue, got)
			}
			if query == nil || query.Page != 1 || query.Size != 10 {
				t.Fatalf("unexpected query: %+v", query)
			}
			return []*model.AWDServiceTemplate{
				{ID: 1, Name: "Bank Portal AWD", Slug: "bank-portal-awd"},
			}, 1, nil
		},
	}
	service := NewAWDServiceTemplateQueryService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.ListTemplatesWithContext(ctx, &dto.AWDServiceTemplateQuery{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListTemplatesWithContext() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected list repository to be called")
	}
	if resp == nil || resp.Total != 1 || len(resp.Items) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
