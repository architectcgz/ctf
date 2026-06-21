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
		"runtime_node_id bigint",
		"ALTER TABLE ONLY public.instances",
		"ADD CONSTRAINT instances_runtime_node_id_fkey",
		"FOREIGN KEY (runtime_node_id) REFERENCES public.runtime_nodes(id)",
		"CREATE INDEX idx_instances_runtime_node_id ON public.instances USING btree (runtime_node_id)",
		"CREATE INDEX idx_instances_runtime_node_container_id ON public.instances USING btree (runtime_node_id, container_id)",
		"WHERE ((runtime_node_id IS NOT NULL) AND ((container_id)::text <> ''::text))",
		"CREATE INDEX idx_instances_runtime_node_network_id ON public.instances USING btree (runtime_node_id, network_id)",
		"WHERE ((runtime_node_id IS NOT NULL) AND ((network_id)::text <> ''::text))",
		"CREATE TABLE public.contest_runtime_placements",
		"runtime_node_id bigint NOT NULL",
		"ADD CONSTRAINT contest_runtime_placements_runtime_node_id_fkey",
		"FOREIGN KEY (runtime_node_id) REFERENCES public.runtime_nodes(id)",
		"CREATE UNIQUE INDEX idx_contest_runtime_placements_active_contest",
		"WHERE ((status)::text = 'active'::text)",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("baseline migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}
	for _, forbidden := range []string{
		"\n    node_id bigint",
		"instances_node_id_fkey",
		"idx_instances_node_id",
	} {
		if strings.Contains(upSQL, forbidden) {
			t.Fatalf("baseline migration should not contain old instance runtime node marker %q", forbidden)
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

func TestInstanceRuntimeNodeIDMigration(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000020_rename_instance_runtime_node_id_and_add_awd_runtime_placements.up.sql"))
	if err != nil {
		t.Fatalf("read instance runtime node id up migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"ALTER TABLE public.instances DROP CONSTRAINT IF EXISTS instances_node_id_fkey",
		"ALTER TABLE public.instances RENAME COLUMN node_id TO runtime_node_id",
		"ALTER INDEX IF EXISTS public.idx_instances_node_id RENAME TO idx_instances_runtime_node_id",
		"ADD CONSTRAINT instances_runtime_node_id_fkey",
		"FOREIGN KEY (runtime_node_id) REFERENCES public.runtime_nodes(id)",
		"CREATE INDEX idx_instances_runtime_node_container_id",
		"ON public.instances USING btree (runtime_node_id, container_id)",
		"WHERE runtime_node_id IS NOT NULL AND container_id <> ''",
		"CREATE INDEX idx_instances_runtime_node_network_id",
		"ON public.instances USING btree (runtime_node_id, network_id)",
		"WHERE runtime_node_id IS NOT NULL AND network_id <> ''",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("instance runtime node id up migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}

	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000020_rename_instance_runtime_node_id_and_add_awd_runtime_placements.down.sql"))
	if err != nil {
		t.Fatalf("read instance runtime node id down migration: %v", err)
	}
	downSQL := string(down)
	for _, snippet := range []string{
		"DROP INDEX IF EXISTS public.idx_instances_runtime_node_container_id",
		"DROP INDEX IF EXISTS public.idx_instances_runtime_node_network_id",
		"ALTER TABLE public.instances DROP CONSTRAINT IF EXISTS instances_runtime_node_id_fkey",
		"ALTER TABLE public.instances RENAME COLUMN runtime_node_id TO node_id",
		"ALTER INDEX IF EXISTS public.idx_instances_runtime_node_id RENAME TO idx_instances_node_id",
		"ADD CONSTRAINT instances_node_id_fkey",
		"FOREIGN KEY (node_id) REFERENCES public.runtime_nodes(id)",
	} {
		if !strings.Contains(downSQL, snippet) {
			t.Fatalf("instance runtime node id down migration should contain %q, got:\n%s", snippet, downSQL)
		}
	}
}

func TestAWDRuntimePlacementMigration(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000020_rename_instance_runtime_node_id_and_add_awd_runtime_placements.up.sql"))
	if err != nil {
		t.Fatalf("read AWD runtime placement up migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"CREATE TABLE public.contest_runtime_placements",
		"contest_id bigint NOT NULL",
		"runtime_node_id bigint NOT NULL",
		"status varchar(16) NOT NULL DEFAULT 'active'",
		"released_at timestamp with time zone",
		"FOREIGN KEY (contest_id) REFERENCES public.contests(id)",
		"FOREIGN KEY (runtime_node_id) REFERENCES public.runtime_nodes(id)",
		"CHECK (status IN ('active', 'released'))",
		"CREATE UNIQUE INDEX idx_contest_runtime_placements_active_contest",
		"ON public.contest_runtime_placements (contest_id)",
		"WHERE status = 'active'",
		"CREATE INDEX idx_contest_runtime_placements_runtime_node_status",
		"ON public.contest_runtime_placements (runtime_node_id, status)",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("AWD runtime placement up migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}

	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000020_rename_instance_runtime_node_id_and_add_awd_runtime_placements.down.sql"))
	if err != nil {
		t.Fatalf("read AWD runtime placement down migration: %v", err)
	}
	downSQL := string(down)
	for _, snippet := range []string{
		"DROP INDEX IF EXISTS public.idx_contest_runtime_placements_runtime_node_status",
		"DROP INDEX IF EXISTS public.idx_contest_runtime_placements_active_contest",
		"DROP TABLE IF EXISTS public.contest_runtime_placements",
	} {
		if !strings.Contains(downSQL, snippet) {
			t.Fatalf("AWD runtime placement down migration should contain %q, got:\n%s", snippet, downSQL)
		}
	}
}

func TestRuntimeNodeLastSeenMigration(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000019_add_runtime_node_last_seen_at.up.sql"))
	if err != nil {
		t.Fatalf("read runtime node last_seen_at up migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"ALTER TABLE public.runtime_nodes",
		"ADD COLUMN last_seen_at timestamp with time zone",
		"idx_runtime_nodes_health_schedulable_seen",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("runtime node last_seen_at up migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}

	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000019_add_runtime_node_last_seen_at.down.sql"))
	if err != nil {
		t.Fatalf("read runtime node last_seen_at down migration: %v", err)
	}
	downSQL := string(down)
	for _, snippet := range []string{
		"DROP INDEX IF EXISTS public.idx_runtime_nodes_health_schedulable_seen",
		"DROP COLUMN IF EXISTS last_seen_at",
	} {
		if !strings.Contains(downSQL, snippet) {
			t.Fatalf("runtime node last_seen_at down migration should contain %q, got:\n%s", snippet, downSQL)
		}
	}
}
