package practice

import (
	"testing"

	runtime "ctf-platform/internal/module/runtime"
)

func TestRuntimeInstanceServiceAdapterContractsCompile(t *testing.T) {
	var _ runtimeInstanceService = (*runtimeInstanceServiceAdapter)(nil)
	_ = NewRuntimeInstanceServiceAdapter((*runtime.Service)(nil))
}
