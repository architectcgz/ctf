package dto

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestConfigureFlagReqRejectsSharedProofFlagType(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/challenges/1/flag",
		bytes.NewBufferString(`{"flag_type":"shared_proof"}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	var req ConfigureFlagReq
	if err := ctx.ShouldBindJSON(&req); err == nil {
		t.Fatal("expected shared_proof flag type to be rejected by request binding")
	}
}
