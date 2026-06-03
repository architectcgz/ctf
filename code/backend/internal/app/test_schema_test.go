package app

import (
	"testing"

	"gorm.io/gorm"

	systemapp "ctf-platform/internal/testutil/systemapp"
)

func openInternalAppTestSQLite(t *testing.T, filename string) *gorm.DB {
	t.Helper()
	return systemapp.OpenInternalAppTestSQLite(t, filename)
}

func reserveInternalAppTestPortRange(t *testing.T, width int) (int, int) {
	t.Helper()
	return systemapp.ReserveInternalAppTestPortRange(t, width)
}

func TestInternalAppTestSchemaIncludesRuntimeTables(t *testing.T) {
	db := openInternalAppTestSQLite(t, "schema-check.sqlite")

	for _, table := range []string{"network_allocations", "runtime_nodes"} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("expected shared internal app test schema to include table %s", table)
		}
	}
}

func TestNewPracticeFlowTestConfigAvoidsOccupiedHostPortRange(t *testing.T) {
	const occupiedPort = 30000

	systemapp.OccupyTestPortIfFree(t, occupiedPort)

	cfg := newPracticeFlowTestConfig(t)
	if cfg.Container.PortRangeStart <= occupiedPort && cfg.Container.PortRangeEnd >= occupiedPort {
		t.Fatalf("expected test config to avoid occupied port %d, got range %d-%d", occupiedPort, cfg.Container.PortRangeStart, cfg.Container.PortRangeEnd)
	}
}
