package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestAccessLogIncludesErrorForServerFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, recorded := observer.New(zap.DebugLevel)
	logger := zap.New(core)

	engine := gin.New()
	engine.Use(AccessLog(logger))
	engine.GET("/api/v1/test", func(c *gin.Context) {
		c.Set(RequestIDKey, "req-test")
		_ = c.Error(errors.New("query failed: relation users does not exist"))
		c.Status(http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != zap.ErrorLevel {
		t.Fatalf("expected error log level, got %s", entry.Level)
	}
	if entry.Message != "http_request" {
		t.Fatalf("expected message http_request, got %s", entry.Message)
	}

	fields := entry.ContextMap()
	if got := fields["error"]; got != "query failed: relation users does not exist" {
		t.Fatalf("expected logged error, got %#v", got)
	}
	if got := fields["request_id"]; got != "req-test" {
		t.Fatalf("expected request id req-test, got %#v", got)
	}
}
