package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/pkg/errcode"
)

func TestFromErrorRecordsAppErrorCause(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	c.Set("request_id", "req-test")

	expected := errors.New("list students by class: column student_no does not exist")
	FromError(c, errcode.ErrInternal.WithCause(expected))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if len(c.Errors) != 1 {
		t.Fatalf("expected one context error, got %d", len(c.Errors))
	}
	if !errors.Is(c.Errors.Last().Err, expected) {
		t.Fatalf("expected recorded error %v, got %v", expected, c.Errors.Last().Err)
	}
}
