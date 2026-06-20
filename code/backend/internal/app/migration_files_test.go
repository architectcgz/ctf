package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImagesDeletedAtIndexInBaseline(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.down.sql"))
	if err != nil {
		t.Fatalf("read baseline down migration: %v", err)
	}
	if !strings.Contains(string(up), "CREATE INDEX idx_images_deleted_at ON public.images USING btree (deleted_at)") {
		t.Fatalf("baseline migration should create idx_images_deleted_at on deleted_at, got:\n%s", string(up))
	}
	if strings.Contains(string(up), "\\restrict ") || strings.Contains(string(up), "\\unrestrict ") {
		t.Fatalf("baseline migration must not contain psql-only restrict directives, got:\n%s", string(up))
	}
	if !strings.Contains(string(down), "DROP SCHEMA IF EXISTS public CASCADE") {
		t.Fatalf("baseline down migration should reset public schema, got:\n%s", string(down))
	}
}

func TestActiveHostPortIndexIgnoresZeroPort(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	upSQL := string(up)
	if !strings.Contains(upSQL, "host_port > 0") {
		t.Fatalf("baseline migration should only enforce active host_port uniqueness for positive ports")
	}
	if strings.Contains(upSQL, "uk_instances_active_host_port ON public.instances USING btree (host_port) WHERE ((host_port IS NOT NULL)") {
		t.Fatalf("baseline migration should not treat host_port=0 as an active published port")
	}
}

func TestBaselineSeedsDefaultLocalUsers(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"INSERT INTO public.roles (code, name, description)",
		"INSERT INTO public.users (",
		"INSERT INTO public.user_roles (user_id, role_id)",
		"'Platform Admin'",
		"'Demo Teacher'",
		"'Demo Student'",
		"'admin'",
		"'teacher'",
		"'student'",
		"'student2'",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("baseline migration should keep default local seed snippet %q", snippet)
		}
	}
}

func TestAWDRuntimeConfigChallengeIDCleanupMigration(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000015_remove_legacy_awd_runtime_config_challenge_id.up.sql"))
	if err != nil {
		t.Fatalf("read AWD runtime config cleanup migration: %v", err)
	}
	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000015_remove_legacy_awd_runtime_config_challenge_id.down.sql"))
	if err != nil {
		t.Fatalf("read AWD runtime config cleanup rollback migration: %v", err)
	}
	upSQL := string(up)
	for _, required := range []string{
		"DO $$",
		"UPDATE public.contest_awd_services",
		"parsed_runtime_config := service_record.runtime_config::jsonb",
		"EXCEPTION WHEN invalid_text_representation THEN",
		"jsonb_typeof(parsed_runtime_config) = 'object'",
		"parsed_runtime_config ? 'challenge_id'",
		"parsed_runtime_config - 'challenge_id'",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("AWD runtime config cleanup migration missing %q:\n%s", required, upSQL)
		}
	}
	if !strings.Contains(string(down), "cannot reconstruct") {
		t.Fatalf("AWD runtime config cleanup rollback should document why it is a no-op:\n%s", string(down))
	}
}

func TestPlatformEventOutboxMigrationQualifiesPublicSchema(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000018_create_platform_event_outbox.up.sql"))
	if err != nil {
		t.Fatalf("read platform event outbox migration: %v", err)
	}
	upSQL := string(up)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS public.platform_event_outbox",
		"ON public.platform_event_outbox",
		"ALTER TABLE public.notifications",
		"ON public.notifications",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("platform event outbox migration must qualify public schema with %q:\n%s", required, upSQL)
		}
	}
}
