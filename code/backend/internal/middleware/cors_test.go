package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/config"
)

func TestCORSEmptyAllowOriginsRejectsPreflight(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(CORS(config.CORSConfig{
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
	}))
	engine.OPTIONS("/api/v1/challenges", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/challenges", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected preflight to be rejected, got %d", resp.Code)
	}
	if origin := resp.Header().Get("Access-Control-Allow-Origin"); origin != "" {
		t.Fatalf("expected no allow-origin header for rejected preflight, got %q", origin)
	}
}

func TestCORSAllowsConfiguredOriginPreflight(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(CORS(config.CORSConfig{
		AllowOrigins:     []string{"https://academy.example.com"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
	}))
	engine.OPTIONS("/api/v1/challenges", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/challenges", nil)
	req.Header.Set("Origin", "https://academy.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected configured origin preflight to pass, got %d", resp.Code)
	}
	if origin := resp.Header().Get("Access-Control-Allow-Origin"); origin != "https://academy.example.com" {
		t.Fatalf("expected configured allow-origin header, got %q", origin)
	}
}
