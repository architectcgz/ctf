package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
)

func TestAWDServiceTemplateHandlerListTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAWDServiceTemplateHandler(
		stubAWDServiceTemplateCommandService{},
		stubAWDServiceTemplateQueryService{
			listWithContextFunc: func(ctx context.Context, req *dto.AWDServiceTemplateQuery) (*dto.AWDServiceTemplatePageResp, error) {
				return &dto.AWDServiceTemplatePageResp{
					Items: []*dto.AWDServiceTemplateResp{
						{ID: 1, Name: "Bank Portal AWD", Slug: "bank-portal-awd"},
					},
					Total: 1,
				}, nil
			},
		},
	)

	ctx, recorder := newJSONTestContext(t, http.MethodGet, "/admin/awd-service-templates", "")
	handler.ListTemplates(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func newJSONTestContext(t *testing.T, method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req
	return ctx, recorder
}

type stubAWDServiceTemplateCommandService struct {
	listImportsFunc func(ctx context.Context, actorUserID int64) ([]dto.AWDServiceTemplateImportPreviewResp, error)
	getImportFunc   func(ctx context.Context, actorUserID int64, id string) (*dto.AWDServiceTemplateImportPreviewResp, error)
}

func (stubAWDServiceTemplateCommandService) CreateTemplate(ctx context.Context, actorUserID int64, req *dto.CreateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

func (stubAWDServiceTemplateCommandService) UpdateTemplate(ctx context.Context, id int64, req *dto.UpdateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

func (stubAWDServiceTemplateCommandService) DeleteTemplate(ctx context.Context, id int64) error {
	return nil
}

func (stubAWDServiceTemplateCommandService) PreviewImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*dto.AWDServiceTemplateImportPreviewResp, error) {
	return nil, nil
}

func (s stubAWDServiceTemplateCommandService) ListImports(ctx context.Context, actorUserID int64) ([]dto.AWDServiceTemplateImportPreviewResp, error) {
	if s.listImportsFunc != nil {
		return s.listImportsFunc(ctx, actorUserID)
	}
	return nil, nil
}

func (s stubAWDServiceTemplateCommandService) GetImport(ctx context.Context, actorUserID int64, id string) (*dto.AWDServiceTemplateImportPreviewResp, error) {
	if s.getImportFunc != nil {
		return s.getImportFunc(ctx, actorUserID, id)
	}
	return nil, nil
}

func (stubAWDServiceTemplateCommandService) CommitImport(ctx context.Context, actorUserID int64, id string) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

type stubAWDServiceTemplateQueryService struct {
	listWithContextFunc func(ctx context.Context, req *dto.AWDServiceTemplateQuery) (*dto.AWDServiceTemplatePageResp, error)
}

func (s stubAWDServiceTemplateQueryService) GetTemplate(ctx context.Context, id int64) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

func (s stubAWDServiceTemplateQueryService) ListTemplates(ctx context.Context, req *dto.AWDServiceTemplateQuery) (*dto.AWDServiceTemplatePageResp, error) {
	if s.listWithContextFunc != nil {
		return s.listWithContextFunc(ctx, req)
	}
	return &dto.AWDServiceTemplatePageResp{}, nil
}

type awdServiceTemplateHandlerContextKey string

func TestAWDServiceTemplateHandlerListImportsPropagatesRequestContextToCommandService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctxKey := awdServiceTemplateHandlerContextKey("list-imports")
	expectedCtxValue := "ctx-list-imports"
	called := false
	handler := NewAWDServiceTemplateHandler(
		stubAWDServiceTemplateCommandService{
			listImportsFunc: func(ctx context.Context, actorUserID int64) ([]dto.AWDServiceTemplateImportPreviewResp, error) {
				called = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected list-imports ctx value %v, got %v", expectedCtxValue, got)
				}
				if actorUserID != 2001 {
					t.Fatalf("unexpected actor user id: %d", actorUserID)
				}
				return []dto.AWDServiceTemplateImportPreviewResp{{ID: "preview-1", Slug: "awd-bank-portal-01"}}, nil
			},
		},
		stubAWDServiceTemplateQueryService{},
	)

	ctx, recorder := newJSONTestContext(t, http.MethodGet, "/admin/awd-service-template-imports", "")
	authctx.SetCurrentUser(ctx, authctx.CurrentUser{UserID: 2001})
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), ctxKey, expectedCtxValue))

	handler.ListImports(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !called {
		t.Fatal("expected list-imports command to be called")
	}
}

func TestAWDServiceTemplateHandlerGetImportPropagatesRequestContextToCommandService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctxKey := awdServiceTemplateHandlerContextKey("get-import")
	expectedCtxValue := "ctx-get-import"
	called := false
	handler := NewAWDServiceTemplateHandler(
		stubAWDServiceTemplateCommandService{
			getImportFunc: func(ctx context.Context, actorUserID int64, id string) (*dto.AWDServiceTemplateImportPreviewResp, error) {
				called = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected get-import ctx value %v, got %v", expectedCtxValue, got)
				}
				if actorUserID != 2001 {
					t.Fatalf("unexpected actor user id: %d", actorUserID)
				}
				if id != "preview-1" {
					t.Fatalf("unexpected import id: %s", id)
				}
				return &dto.AWDServiceTemplateImportPreviewResp{ID: id, Slug: "awd-bank-portal-01"}, nil
			},
		},
		stubAWDServiceTemplateQueryService{},
	)

	ctx, recorder := newJSONTestContext(t, http.MethodGet, "/admin/awd-service-template-imports/preview-1", "")
	ctx.Params = gin.Params{{Key: "id", Value: "preview-1"}}
	authctx.SetCurrentUser(ctx, authctx.CurrentUser{UserID: 2001})
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), ctxKey, expectedCtxValue))

	handler.GetImport(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !called {
		t.Fatal("expected get-import command to be called")
	}
}
