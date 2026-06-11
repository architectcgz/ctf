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
	practiceports "ctf-platform/internal/module/practice/ports"
)

type stubPracticeProgressQuery struct {
	getProgressFn func(ctx context.Context, userID int64) (*practiceports.UserProgressSnapshot, error)
	getTimelineFn func(ctx context.Context, userID int64, limit, offset int) (*practiceports.TimelineSnapshot, error)
}

func (s *stubPracticeProgressQuery) GetProgress(ctx context.Context, userID int64) (*practiceports.UserProgressSnapshot, error) {
	return s.getProgressFn(ctx, userID)
}

func (s *stubPracticeProgressQuery) GetTimeline(ctx context.Context, userID int64, limit, offset int) (*practiceports.TimelineSnapshot, error) {
	return s.getTimelineFn(ctx, userID, limit, offset)
}

func TestHandlerUsesPracticeQueryForProgress(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil, nil, nil, nil, &stubPracticeProgressQuery{
		getProgressFn: func(ctx context.Context, userID int64) (*practiceports.UserProgressSnapshot, error) {
			if userID != 42 {
				t.Fatalf("unexpected user id: %d", userID)
			}
			return &practiceports.UserProgressSnapshot{TotalScore: 120}, nil
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
		Code int          `json:"code"`
		Data ProgressResp `json:"data"`
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

	now := time.Now().UTC()
	handler := NewHandler(nil, nil, nil, nil, &stubPracticeProgressQuery{
		getTimelineFn: func(ctx context.Context, userID int64, limit, offset int) (*practiceports.TimelineSnapshot, error) {
			if userID != 7 {
				t.Fatalf("unexpected user id: %d", userID)
			}
			if limit != 5 || offset != 2 {
				t.Fatalf("unexpected pagination: limit=%d offset=%d", limit, offset)
			}
			return &practiceports.TimelineSnapshot{
				Events: []practiceports.TimelineEventSnapshot{{
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

	var envelope struct {
		Code int          `json:"code"`
		Data TimelineResp `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got := len(envelope.Data.Events); got != 1 {
		t.Fatalf("expected 1 event, got %d", got)
	}
}
