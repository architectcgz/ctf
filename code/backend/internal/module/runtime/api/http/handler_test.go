package http

import (
	"testing"

	runtime "ctf-platform/internal/module/runtime"
)

func TestHandlerContractsCompile(t *testing.T) {
	var _ runtimeService = (*runtime.Module)(nil)
	_ = NewHandler((*runtime.Module)(nil), nil, CookieConfig{})
}
