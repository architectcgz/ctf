package shared

import (
	"testing"

	"ctf-platform/internal/testutil/archtest"
)

func TestSharedPackagesStayLightweightAndModuleAgnostic(t *testing.T) {
	t.Parallel()

	for _, file := range archtest.RuntimeGoFiles(t, ".") {
		for _, importPath := range archtest.Imports(t, file) {
			for _, blocked := range forbiddenSharedRuntimeImports {
				if archtest.ImportPathMatches(importPath, blocked) {
					t.Fatalf("%s must not import %s; shared packages should stay lightweight and module-agnostic", file, importPath)
				}
			}
		}
	}
}

var forbiddenSharedRuntimeImports = []string{
	"ctf-platform/internal/app",
	"ctf-platform/internal/bootstrap",
	"ctf-platform/internal/config",
	"ctf-platform/internal/dto",
	"ctf-platform/internal/infrastructure",
	"ctf-platform/internal/middleware",
	"ctf-platform/internal/model",
	"ctf-platform/internal/module",
	"ctf-platform/internal/platform",
	"ctf-platform/internal/validation",
	"database/sql",
	"github.com/docker/docker",
	"github.com/gin-gonic/gin",
	"github.com/redis/go-redis",
	"gorm.io/gorm",
	"net/http",
}
