package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/platform/requestctx"
)

func TestRequestIDWritesIntoRequestContext(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	RequestID()(ctx)

	requestID := ctx.GetString(RequestIDKey)
	if requestID == "" {
		t.Fatal("expected request id in gin context")
	}
	if got := requestctx.RequestIDFromContext(ctx.Request.Context()); got != requestID {
		t.Fatalf("request context request_id = %q, want %q", got, requestID)
	}
}

func TestRequestIDCanonicalizesIncomingHeader(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", " req-123 ")
	ctx.Request = req

	RequestID()(ctx)

	const want = "req-123"
	if got := ctx.GetString(RequestIDKey); got != want {
		t.Fatalf("gin context request_id = %q, want %q", got, want)
	}
	if got := requestctx.RequestIDFromContext(ctx.Request.Context()); got != want {
		t.Fatalf("request context request_id = %q, want %q", got, want)
	}
	if got := ctx.Request.Header.Get("X-Request-ID"); got != want {
		t.Fatalf("request header request_id = %q, want %q", got, want)
	}
	if got := recorder.Header().Get("X-Request-ID"); got != want {
		t.Fatalf("response header request_id = %q, want %q", got, want)
	}
}

func TestRequestIDBlankHeaderFallsBackToGeneratedValue(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "   ")
	ctx.Request = req

	RequestID()(ctx)

	requestID := ctx.GetString(RequestIDKey)
	if requestID == "" {
		t.Fatal("expected generated request id")
	}
	if got := requestctx.RequestIDFromContext(ctx.Request.Context()); got != requestID {
		t.Fatalf("request context request_id = %q, want %q", got, requestID)
	}
	if got := ctx.Request.Header.Get("X-Request-ID"); got != requestID {
		t.Fatalf("request header request_id = %q, want %q", got, requestID)
	}
	if got := recorder.Header().Get("X-Request-ID"); got != requestID {
		t.Fatalf("response header request_id = %q, want %q", got, requestID)
	}
}
