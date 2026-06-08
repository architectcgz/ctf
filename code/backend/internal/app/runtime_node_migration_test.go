package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeNodeContractInBaseline(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"CREATE TABLE public.runtime_nodes",
		"CREATE TABLE public.instances",
		"node_id bigint",
		"ALTER TABLE ONLY public.instances",
		"ADD CONSTRAINT instances_node_id_fkey",
		"FOREIGN KEY (node_id) REFERENCES public.runtime_nodes(id)",
		"CREATE INDEX idx_instances_node_id ON public.instances USING btree (node_id)",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("baseline migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}

	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.down.sql"))
	if err != nil {
		t.Fatalf("read baseline down migration: %v", err)
	}
	if !strings.Contains(string(down), "DROP SCHEMA IF EXISTS public CASCADE") {
		t.Fatalf("baseline down migration should reset public schema, got:\n%s", string(down))
	}
}
