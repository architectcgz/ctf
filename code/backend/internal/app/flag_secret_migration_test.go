package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlagSecretClusterContractInBaseline(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}

	upSQL := string(up)
	for _, required := range []string{
		"CREATE TABLE public.runtime_cluster_secrets",
		"active_key_id",
		"active_fingerprint",
		"key_fingerprints",
		"flag_key_id character varying(128)",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("baseline flag secret contract missing %q:\n%s", required, upSQL)
		}
	}
}
