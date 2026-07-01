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
		"INSERT INTO public.roles VALUES",
		"INSERT INTO public.users VALUES",
		"INSERT INTO public.user_roles VALUES",
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

func TestEventOutboxContractsInBaseline(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	upSQL := string(up)
	for _, required := range []string{
		"CREATE TABLE public.contest_realtime_outbox",
		"CREATE UNIQUE INDEX uk_contest_realtime_outbox_dedupe_key ON public.contest_realtime_outbox",
		"CREATE TABLE public.platform_event_outbox",
		"ON public.platform_event_outbox",
		"source_event_key text DEFAULT ''::text NOT NULL",
		"ON public.notifications",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("baseline event outbox contract missing %q:\n%s", required, upSQL)
		}
	}
}

func TestOAuthBrowserAuthorizationMigrationFiles(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000002_oauth_browser_authorization.up.sql"))
	if err != nil {
		t.Fatalf("read oauth migration: %v", err)
	}
	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000002_oauth_browser_authorization.down.sql"))
	if err != nil {
		t.Fatalf("read oauth down migration: %v", err)
	}

	upSQL := string(up)
	for _, required := range []string{
		"CREATE TABLE public.oauth_clients",
		"client_id text NOT NULL",
		"redirect_uris jsonb NOT NULL",
		"token_endpoint_auth_method text DEFAULT 'none'::text NOT NULL",
		"CREATE TABLE public.oauth_consents",
		"user_id bigint NOT NULL",
		"client_id text NOT NULL",
		"scope text NOT NULL",
		"UNIQUE (user_id, client_id, scope)",
		"CREATE INDEX idx_oauth_consents_user_client",
		"REFERENCES public.users(id) ON DELETE CASCADE",
		"REFERENCES public.oauth_clients(client_id) ON DELETE CASCADE",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("oauth migration should contain %q, got:\n%s", required, upSQL)
		}
	}

	downSQL := string(down)
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS public.oauth_consents") ||
		!strings.Contains(downSQL, "DROP TABLE IF EXISTS public.oauth_clients") {
		t.Fatalf("oauth down migration should drop consent and client tables, got:\n%s", downSQL)
	}
}
