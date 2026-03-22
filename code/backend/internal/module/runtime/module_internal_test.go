package runtime

import (
	"testing"

	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
)

func TestModuleRepositoryContractsCompile(t *testing.T) {
	var _ instanceRepository = (*Module)(nil)
	var _ instanceRepository = (*runtimeinfrarepo.Repository)(nil)
}
