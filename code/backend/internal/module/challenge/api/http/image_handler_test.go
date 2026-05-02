package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	"ctf-platform/internal/validation"
)

type stubImageCommandService struct {
	createCalled bool
}

func (s *stubImageCommandService) CreateImage(ctx context.Context, req challengecmd.CreateImageInput) (*dto.ImageResp, error) {
	s.createCalled = true
	return &dto.ImageResp{}, nil
}

func (s *stubImageCommandService) UpdateImage(ctx context.Context, id int64, req challengecmd.UpdateImageInput) error {
	return nil
}

func (s *stubImageCommandService) DeleteImage(ctx context.Context, id int64) error {
	return nil
}

type stubImageQueryService struct{}

func (s stubImageQueryService) GetImage(ctx context.Context, id int64) (*dto.ImageResp, error) {
	return &dto.ImageResp{}, nil
}

func (s stubImageQueryService) ListImages(ctx context.Context, query *dto.ImageQuery) (*dto.PageResult[*dto.ImageResp], error) {
	return &dto.PageResult[*dto.ImageResp]{}, nil
}

func TestImageHandlerCreateImageRejectsInvalidImageName(t *testing.T) {
	t.Parallel()

	if err := validation.Register(); err != nil {
		t.Fatalf("register validator: %v", err)
	}
	gin.SetMode(gin.TestMode)

	commands := &stubImageCommandService{}
	handler := NewImageHandler(commands, stubImageQueryService{})
	router := gin.New()
	router.POST("/images", handler.CreateImage)

	req := httptest.NewRequest(http.MethodPost, "/images", strings.NewReader(`{"name":"CTF/Web;rm -rf /","tag":"latest"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got status=%d body=%s", rec.Code, rec.Body.String())
	}
	if commands.createCalled {
		t.Fatal("expected invalid image name to be rejected before command service")
	}
}
