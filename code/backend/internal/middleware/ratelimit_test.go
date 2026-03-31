package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"

	ratelimitpkg "ctf-platform/pkg/ratelimit"
)

func TestRateLimitByLoginPrincipalAndIP_AllowsDifferentPrincipalsFromSameIP(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mini.Close()

	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	defer func() {
		_ = cache.Close()
	}()
	if err := cache.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}

	checker := ratelimitTestChecker(cache)
	router := gin.New()
	router.POST("/login",
		RateLimitByLoginPrincipalAndIP(checker, "login", 1, time.Minute),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	first := performRateLimitJSONRequest(t, router, "/login", map[string]any{"username": "alice", "password": "secret"}, "10.0.0.1")
	if first.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want %d", first.Code, http.StatusOK)
	}

	second := performRateLimitJSONRequest(t, router, "/login", map[string]any{"username": "bob", "password": "secret"}, "10.0.0.1")
	if second.Code != http.StatusOK {
		t.Fatalf("second request status = %d, want %d", second.Code, http.StatusOK)
	}

	third := performRateLimitJSONRequest(t, router, "/login", map[string]any{"username": "alice", "password": "secret"}, "10.0.0.1")
	if third.Code != http.StatusTooManyRequests {
		t.Fatalf("third request status = %d, want %d", third.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimitByLoginPrincipalAndIP_PreservesRequestBodyForHandler(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mini.Close()

	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	defer func() {
		_ = cache.Close()
	}()
	if err := cache.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}

	checker := ratelimitTestChecker(cache)
	router := gin.New()
	router.POST("/login",
		RateLimitByLoginPrincipalAndIP(checker, "login", 2, time.Minute),
		func(c *gin.Context) {
			var payload struct {
				Username string `json:"username"`
			}
			if err := c.ShouldBindJSON(&payload); err != nil {
				c.Status(http.StatusBadRequest)
				return
			}
			c.String(http.StatusOK, payload.Username)
		},
	)

	resp := performRateLimitJSONRequest(t, router, "/login", map[string]any{"username": "alice", "password": "secret"}, "10.0.0.1")
	if resp.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d", resp.Code, http.StatusOK)
	}
	if body := strings.TrimSpace(resp.Body.String()); body != "alice" {
		t.Fatalf("response body = %q, want %q", body, "alice")
	}
}

func ratelimitTestChecker(cache *redislib.Client) *ratelimitpkg.Checker {
	return ratelimitpkg.NewChecker(cache, "test:ratelimit")
}

func performRateLimitJSONRequest(t *testing.T, router http.Handler, path string, payload map[string]any, ip string) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = ip + ":12345"

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}
