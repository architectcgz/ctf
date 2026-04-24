package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type awdServiceTemplateQueryContextRepoStub struct {
	findByIDFn func(ctx context.Context, id int64) (*model.AWDServiceTemplate, error)
	listFn     func(ctx context.Context, query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error)
}

func (s *awdServiceTemplateQueryContextRepoStub) FindAWDServiceTemplateByID(ctx context.Context, id int64) (*model.AWDServiceTemplate, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *awdServiceTemplateQueryContextRepoStub) ListAWDServiceTemplates(ctx context.Context, query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, query)
	}
	return nil, 0, nil
}

type awdServiceTemplateQueryContextKey string

func TestAWDServiceTemplateQueryServiceGetTemplatePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdServiceTemplateQueryContextKey("get")
	expectedCtxValue := "ctx-get"
	findCalled := false
	repo := &awdServiceTemplateQueryContextRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.AWDServiceTemplate, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.AWDServiceTemplate{ID: id, Name: "Bank Portal AWD", Slug: "bank-portal-awd"}, nil
		},
	}
	service := NewAWDServiceTemplateQueryService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetTemplate(ctx, 77)
	if err != nil {
		t.Fatalf("GetTemplate() error = %v", err)
	}
	if !findCalled {
		t.Fatal("expected find repository to be called")
	}
	if resp == nil || resp.ID != 77 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAWDServiceTemplateQueryServiceListTemplatesPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdServiceTemplateQueryContextKey("list")
	expectedCtxValue := "ctx-list"
	listCalled := false
	repo := &awdServiceTemplateQueryContextRepoStub{
		listFn: func(ctx context.Context, query *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error) {
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
	resp, err := service.ListTemplates(ctx, &dto.AWDServiceTemplateQuery{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected list repository to be called")
	}
	if resp == nil || resp.Total != 1 || len(resp.Items) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
