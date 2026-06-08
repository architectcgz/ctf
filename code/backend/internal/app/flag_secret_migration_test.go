package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlagSecretClusterContractMigration(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000013_add_runtime_cluster_secret_and_flag_key_id.up.sql"))
	if err != nil {
		t.Fatalf("read flag secret migration: %v", err)
	}
	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000013_add_runtime_cluster_secret_and_flag_key_id.down.sql"))
	if err != nil {
		t.Fatalf("read flag secret rollback migration: %v", err)
	}

	upSQL := string(up)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS public.runtime_cluster_secrets",
		"active_key_id",
		"active_fingerprint",
		"key_fingerprints",
		"ALTER TABLE public.instances ADD COLUMN IF NOT EXISTS flag_key_id",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("flag secret migration missing %q:\n%s", required, upSQL)
		}
	}

	downSQL := string(down)
	for _, required := range []string{
		"ALTER TABLE public.instances DROP COLUMN IF EXISTS flag_key_id",
		"DROP TABLE IF EXISTS public.runtime_cluster_secrets",
	} {
		if !strings.Contains(downSQL, required) {
			t.Fatalf("flag secret rollback missing %q:\n%s", required, downSQL)
		}
	}
}
