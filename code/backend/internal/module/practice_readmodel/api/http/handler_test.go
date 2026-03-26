package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	practicereadmodelqueries "ctf-platform/internal/module/practice_readmodel/application/queries"
)

type stubPracticeQuery struct {
	getProgressFn func(ctx context.Context, userID int64) (*dto.ProgressResp, error)
	getTimelineFn func(ctx context.Context, userID int64, limit, offset int) (*dto.TimelineResp, error)
}

func (s *stubPracticeQuery) GetProgress(ctx context.Context, userID int64) (*dto.ProgressResp, error) {
	return s.getProgressFn(ctx, userID)
}

func (s *stubPracticeQuery) GetTimeline(ctx context.Context, userID int64, limit, offset int) (*dto.TimelineResp, error) {
	return s.getTimelineFn(ctx, userID, limit, offset)
}

func TestNewHandlerAcceptsPracticeQueryContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var _ practicereadmodelqueries.Service = (*stubPracticeQuery)(nil)

	handler := NewHandler(&stubPracticeQuery{
		getProgressFn: func(ctx context.Context, userID int64) (*dto.ProgressResp, error) {
			if userID != 42 {
				t.Fatalf("unexpected user id: %d", userID)
			}
			return &dto.ProgressResp{TotalScore: 120}, nil
		},
	})

	router := gin.New()
	router.GET("/progress", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 42})
		handler.GetProgress(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/progress", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var envelope struct {
		Code int              `json:"code"`
		Data dto.ProgressResp `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data.TotalScore != 120 {
		t.Fatalf("unexpected total score: %+v", envelope.Data)
	}
}

func TestHandlerUsesPracticeQueryForTimeline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()
	handler := NewHandler(&stubPracticeQuery{
		getTimelineFn: func(ctx context.Context, userID int64, limit, offset int) (*dto.TimelineResp, error) {
			if userID != 7 {
				t.Fatalf("unexpected user id: %d", userID)
			}
			if limit != 5 || offset != 2 {
				t.Fatalf("unexpected pagination: limit=%d offset=%d", limit, offset)
			}
			return &dto.TimelineResp{
				Events: []dto.TimelineEvent{{
					Type:      "flag_submit",
					Timestamp: now,
					Detail:    "ok",
				}},
			}, nil
		},
	})

	router := gin.New()
	router.GET("/timeline", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 7})
		handler.GetTimeline(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/timeline?limit=5&offset=2", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
}
