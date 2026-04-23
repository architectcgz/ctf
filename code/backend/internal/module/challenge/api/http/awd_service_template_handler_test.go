package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

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

type stubAWDServiceTemplateCommandService struct{}

func (stubAWDServiceTemplateCommandService) CreateTemplate(actorUserID int64, req *dto.CreateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

func (stubAWDServiceTemplateCommandService) CreateTemplateWithContext(ctx context.Context, actorUserID int64, req *dto.CreateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

func (stubAWDServiceTemplateCommandService) UpdateTemplate(id int64, req *dto.UpdateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

func (stubAWDServiceTemplateCommandService) UpdateTemplateWithContext(ctx context.Context, id int64, req *dto.UpdateAWDServiceTemplateReq) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

func (stubAWDServiceTemplateCommandService) DeleteTemplate(id int64) error {
	return nil
}

func (stubAWDServiceTemplateCommandService) DeleteTemplateWithContext(ctx context.Context, id int64) error {
	return nil
}

func (stubAWDServiceTemplateCommandService) PreviewImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*dto.AWDServiceTemplateImportPreviewResp, error) {
	return nil, nil
}

func (stubAWDServiceTemplateCommandService) ListImports(actorUserID int64) ([]dto.AWDServiceTemplateImportPreviewResp, error) {
	return nil, nil
}

func (stubAWDServiceTemplateCommandService) GetImport(actorUserID int64, id string) (*dto.AWDServiceTemplateImportPreviewResp, error) {
	return nil, nil
}

func (stubAWDServiceTemplateCommandService) CommitImport(ctx context.Context, actorUserID int64, id string) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

type stubAWDServiceTemplateQueryService struct {
	listWithContextFunc func(ctx context.Context, req *dto.AWDServiceTemplateQuery) (*dto.AWDServiceTemplatePageResp, error)
}

func (s stubAWDServiceTemplateQueryService) GetTemplate(id int64) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

func (s stubAWDServiceTemplateQueryService) GetTemplateWithContext(ctx context.Context, id int64) (*dto.AWDServiceTemplateResp, error) {
	return nil, nil
}

func (s stubAWDServiceTemplateQueryService) ListTemplates(req *dto.AWDServiceTemplateQuery) (*dto.AWDServiceTemplatePageResp, error) {
	return &dto.AWDServiceTemplatePageResp{}, nil
}

func (s stubAWDServiceTemplateQueryService) ListTemplatesWithContext(ctx context.Context, req *dto.AWDServiceTemplateQuery) (*dto.AWDServiceTemplatePageResp, error) {
	if s.listWithContextFunc != nil {
		return s.listWithContextFunc(ctx, req)
	}
	return &dto.AWDServiceTemplatePageResp{}, nil
}
