<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# 退役 AWD runtime_config 历史兼容层 Implementation Plan

**Goal:** Retire the historical `contest_awd_services.runtime_config.challenge_id` compatibility layer.

**Architecture:** `contest_awd_services.id` remains the AWD runtime identity, and `contest_awd_services.awd_challenge_id` remains the formal service-to-challenge relation. Historical `runtime_config.challenge_id` values are cleaned by a formal SQL migration, then response/query mappers stop filtering that key because it is no longer part of supported storage or API compatibility.

**Tech Stack:** Go, PostgreSQL SQL migrations, GORM-backed repository tests, CTF backend architecture guards

---

## Task Metadata

- Task Slug: `2026-06-10-awd-runtime-config-compat-retirement`
- Started At: `2026-06-10T11:06:33Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-10-awd-runtime-config-compat-retirement`
- Branch: `task/2026-06-10-awd-runtime-config-compat-retirement`

## Objective And Non-Goals

- Objective:
  - Add a forward migration that removes `challenge_id` from persisted `contest_awd_services.runtime_config` JSON payloads.
  - Remove the command/query response mapper sanitizer that existed only to hide historical `runtime_config.challenge_id`.
  - Update active architecture / todo docs so the compatibility layer is no longer recorded as open migration debt.
- Non-Goals:
  - Do not remove `contest_awd_services.awd_challenge_id`.
  - Do not remove `awd_team_services.awd_challenge_id`, `awd_attack_logs.awd_challenge_id`, or `awd_traffic_events.awd_challenge_id`; these remain display / compatibility fields by current architecture.
  - Do not change checker preview create-before-service behavior where `awd_challenge_id` is still a valid input.
  - Do not solve the separate `challenges.image_id = 0` sentinel or assessment training-image semantics debt.

## Inputs

- Source docs:
  - `docs/文档规范.md`
  - `docs/architecture/backend/design/awd-engine-migration.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
  - `docs/reviews/backend/2026-06-10-gate-review-awd-runtime-config-compat-retirement.md`
- Related architecture/contracts:
  - `code/backend/migrations/`
  - `code/backend/internal/module/contest/application/commands/response_mappers.go`
  - `code/backend/internal/module/contest/application/queries/contest_awd_service_query.go`
  - `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
- Related prior work:
  - Existing create/update tests already prove new writes do not persist `runtime_config.challenge_id`.

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - Changes persisted-data migration behavior and removes a compatibility branch from backend response mapping.
  - Updates active architecture and migration-debt documentation.

## Files

- Create:
  - `code/backend/migrations/000015_remove_legacy_awd_runtime_config_challenge_id.up.sql`
  - `code/backend/migrations/000015_remove_legacy_awd_runtime_config_challenge_id.down.sql`
- Modify:
  - `code/backend/internal/app/migration_files_test.go`
  - `code/backend/internal/module/contest/application/queries/contest_awd_service_query.go`
  - `code/backend/internal/module/contest/application/queries/contest_awd_service_query_test.go`
  - `code/backend/internal/module/contest/application/commands/response_mappers.go`
  - `docs/architecture/backend/design/awd-engine-migration.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Review:
  - `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
- Test:
  - `code/backend/internal/app/migration_files_test.go`
  - `code/backend/internal/module/contest/application/queries/contest_awd_service_query_test.go`
  - `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - Current migration numbering in `code/backend/migrations`.
  - Existing migration guard tests in `code/backend/internal/app/*migration*_test.go`.
  - Existing mapper sanitizer functions and create/update tests that already block new writes.
- Reuse / extend / split / create-new decision:
  - Create a new `000015` migration instead of editing the squashed baseline, because `000013` and `000014` already exist after baseline and this is a forward cleanup for installed databases.
  - Extend existing migration guard tests rather than adding a new harness.
- Owner boundary:
  - PostgreSQL migration owns persisted historical data cleanup.
  - Contest command/query mappers own response shaping and should no longer contain historical field filtering.
  - AWD architecture doc owns the stable `service_id` / `awd_challenge_id` / runtime config contract.
- Why this is the narrowest safe surface:
  - It removes exactly the retired `runtime_config.challenge_id` compatibility branch and leaves all current runtime identity fields unchanged.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The task changes compatibility, data cleanup, docs, and backend response behavior; code and docs needed to establish the owner boundary.
- grill-with-docs findings:
  - The ambiguous term "历史兼容层" resolves to `contest_awd_services.runtime_config.challenge_id`, not the broader `challenge_id` display fields.
  - Removing the mapper sanitizer without a migration would re-expose dirty historical data; the migration must come first in the change.
  - Runtime fact tables' `challenge_id` display fields remain out of scope.
- Plan adjustments after challenge:
  - Add `000015` data migration.
  - Keep existing new-write tests.
  - Remove only mapper-level compatibility filtering and docs backlog item.

## Validation

- Commands:
  - `cd code/backend && go generate ./internal/module/contest/application/commands`
  - `cd code/backend && go test ./internal/app -run 'TestAWDRuntimeConfigChallengeIDCleanupMigration|TestImagesDeletedAtIndexInBaseline|TestActiveHostPortIndexIgnoresZeroPort|TestBaselineSeedsDefaultLocalUsers' -count=1`
  - `cd code/backend && go test ./internal/module/contest/application/queries -run TestContestAWDServiceQueryServiceListContestAWDServicesIncludesValidationState -count=1`
  - `cd code/backend && go test ./internal/module/contest/application/commands -run 'TestContestAWDServiceService(Create|Update)DoesNotPersistLegacyChallengeIDInRuntimeConfig' -count=1`
  - `cd code/backend && go test ./internal/module/contest/application/queries ./internal/module/contest/application/commands -count=1`
  - `cd code/backend && go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
  - `python3 scripts/check-docs-consistency.py`
  - `bash scripts/check-startup-gate.sh`
  - `bash scripts/check-workflow-governance.sh`
  - `git diff --check`
  - `docker exec ctf-postgres psql -U postgres -d ctf -v ON_ERROR_STOP=1 -P pager=off -c "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1;"`
  - `timeout 120 go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3 -path ./migrations -database 'postgres://postgres:postgres123456@127.0.0.1:15432/ctf?sslmode=disable' up`
  - `docker exec ctf-postgres psql -U postgres -d ctf -v ON_ERROR_STOP=1 -P pager=off -c "SELECT COUNT(*) AS candidate_rows FROM public.contest_awd_services WHERE runtime_config IS NOT NULL AND btrim(runtime_config) <> '' AND runtime_config LIKE '%challenge_id%';"`
  - `docker exec ctf-postgres psql -U postgres -d ctf -v ON_ERROR_STOP=1 -P pager=off -c "SELECT indexname, indexdef FROM pg_indexes WHERE schemaname = 'public' AND tablename = 'instances' AND indexname = 'idx_instances_status_updated_id';"`
- Manual checks:
  - `rg -n "sanitizeContestAWDServiceRuntimeConfig|runtime_config.challenge_id|compatibility challenge_id" code/backend/internal/module/contest docs/architecture/backend/design/awd-engine-migration.md docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Review focus:
  - Migration must not alter current `awd_challenge_id` relations.
  - Mapper changes must not drop current checker/runtime metadata.
  - Docs must distinguish retired `runtime_config.challenge_id` from retained display `challenge_id` fields.

## Live Database Migration Evidence

- Target: local `ctf-postgres`, database `ctf`
- Backup: `/home/azhi/workspace/projects/.backups/ctf-postgres/ctf-before-000014-000015-20260610-193347.sql.gz`
  - Verified with `gzip -t`
  - Restore shape: `gzip -dc <backup> | docker exec -i ctf-postgres psql -U postgres -d <target-db>`
- Before migration: `schema_migrations.version = 13`, `dirty = false`
- Applied:
  - `14/u instances_stopping_cleanup_index`
  - `15/u remove_legacy_awd_runtime_config_challenge_id`
- After migration:
  - `schema_migrations.version = 15`, `dirty = false`
  - `contest_awd_services` candidate rows containing `challenge_id`: `0`
  - `idx_instances_status_updated_id` exists on `public.instances(status, updated_at, id)`
