# Backend Migration Squash Implementation Plan

**Goal:** 把 `code/backend/migrations` 从 `000001..000012` 的增量链收口成单份 baseline，同时保留一份可提交的单文件 schema SQL 源码，并把旧链路里直到 `000012` 才补齐的 schema / replay 问题统一吸收到当前 baseline。

**Architecture:** 正式 runtime schema owner 仍然是 `code/backend/migrations/000001_init_schema.up.sql` 与 `golang-migrate`。`code/backend/schema/ctf_schema_submission.sql` 只是提交 / 参考用 schema snapshot，不参与 runtime 迁移。

**Tech Stack:** PostgreSQL, golang-migrate, SQL migrations, pg_dump, Go migration guardrail tests, docs

---

## Task Metadata

- Task Slug: `2026-06-08-backend-migration-squash`
- Started At: `2026-06-08T00:00:00Z`
- Worktree: `/home/azhi/workspace/projects/ctf`
- Branch: `main`

## Objective And Non-Goals

- Objective:
  - 生成并保留一份可提交的单文件 schema SQL。
  - 用当前完整 schema 替换 `000001_init_schema.up.sql` baseline。
  - 删除 `000002..000012` 的 runtime migration 链，收口成单 baseline。
  - 修正 / 改写直接锁旧增量文件号的 migration 测试和活跃文档引用。
- Non-Goals:
  - 不保留旧增量 migration 作为并行 runtime owner。
  - 不引入新的 migration 工具或 sidecar。
  - 不处理历史 Git 追溯以外的额外 archive 方案。

## Inputs

- Source docs:
  - `docs/architecture/backend/02-database-design.md`
  - `docs/architecture/backend/design/awd-engine-migration.md`
- Related architecture/contracts:
  - `code/backend/migrations/*.sql`
  - `code/backend/schema/ctf_schema_submission.sql`
  - `code/backend/internal/app/*migration*_test.go`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-05-ctf-api-auto-migrate-implementation-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 直接重定义 runtime migration baseline，触达数据库 owner、测试护栏和架构事实文档。
  - 包含高风险删除：`code/backend/migrations/000002..000012`。

## Files

- Modify:
  - `code/backend/migrations/000001_init_schema.up.sql`
  - `code/backend/migrations/000001_init_schema.down.sql`
  - `code/backend/internal/app/migration_files_test.go`
  - `code/backend/internal/app/contest_status_transition_migration_test.go`
  - `code/backend/internal/app/contest_paused_seconds_migration_test.go`
  - `code/backend/internal/app/runtime_node_migration_test.go`
  - `docs/architecture/backend/02-database-design.md`
  - `docs/architecture/backend/design/awd-engine-migration.md`
- Create:
  - `code/backend/schema/ctf_schema_submission.sql`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-08-backend-migration-squash-implementation-plan.md`
- Delete:
  - `code/backend/migrations/000002_create_awd_service_operations.{up,down}.sql`
  - `code/backend/migrations/000003_create_contest_status_transitions.{up,down}.sql`
  - `code/backend/migrations/000004_fix_active_host_port_index.{up,down}.sql`
  - `code/backend/migrations/000005_create_image_build_jobs.{up,down}.sql`
  - `code/backend/migrations/000006_create_awd_defense_workspaces.{up,down}.sql`
  - `code/backend/migrations/000007_add_contest_paused_seconds.{up,down}.sql`
  - `code/backend/migrations/000008_create_awd_scope_controls.{up,down}.sql`
  - `code/backend/migrations/000009_create_network_allocations.{up,down}.sql`
  - `code/backend/migrations/000010_add_low_risk_foreign_keys.{up,down}.sql`
  - `code/backend/migrations/000011_add_additional_foreign_keys.{up,down}.sql`
  - `code/backend/migrations/000012_create_runtime_nodes.{up,down}.sql`

## Ordered Task Slices

1. Baseline rewrite
   - Replace `000001_init_schema.up.sql` with the sanitized full-schema snapshot.
   - Keep `000001_init_schema.down.sql` as the full public schema reset path.
2. Guardrail rewrite
   - Move old incremental migration assertions onto the new baseline contract.
   - Keep a dedicated check that the baseline still covers runtime nodes, contest status transitions, paused seconds, and host-port uniqueness semantics.
3. Chain squash
   - Remove `000002..000012` from runtime migrations.
   - Update active docs that still cite deleted file numbers.
4. Verification
   - Recreate a temporary database, run the squashed baseline from clean state, and export the submission snapshot.
   - Run the narrow Go migration tests and project workflow completion gate.

## Data, Migration, And Compatibility Impact

- New clean databases will bootstrap from a single baseline instead of replaying 12 migrations.
- Databases whose `schema_migrations` state still points at the removed `000002..000012` chain are no longer upgraded in place, even if their schema shape already matches the current baseline; local development and docker-compose startup paths must reset or recreate those databases before rerunning migrations.
- Historical step-by-step replay from `000002..000012` is intentionally removed from runtime owner paths and remains recoverable only through Git history.

## 复用与 Owner 决策

- Reuse:
  - 继续复用 `golang-migrate` 作为唯一 runtime schema owner，不引入新的 sidecar 或应用启动期 schema owner。
  - 复用现有 migration guardrail tests，把断言目标从旧增量文件改到当前 baseline 合约。
- Owner:
  - `code/backend/migrations/000001_init_schema.up.sql` 负责正式 schema 真相与默认本地 seed。
  - `code/backend/schema/ctf_schema_submission.sql` 只负责提交态 / 参考态 schema snapshot，不参与 runtime replay。
  - `code/backend/scripts/dev-run.sh` 与 `code/backend/scripts/docker-entrypoint.sh` 只负责触发 migration 与旧链 reset 提示，不承担 schema 变更逻辑。
- Why:
  - 这条边界能同时去掉隐式 `AutoMigrate` 漂移、缩短 fresh DB replay 路径，并把“旧链不再支持”的行为明确收口到脚本提示和 baseline replay 合约里。

## Intake Analysis Gate

- Analysis skill:
  - 任务属于非琐碎后端 / migration owner 收口，先按 `code-workflow` 进入 task-slug + implementation-plan 路径。
  - 核心分析问题是 baseline owner、旧链兼容策略、默认本地 seed 契约，以及活跃文档 / 测试是否仍锁定旧 migration 文件号。
- Grill result:
  - 确认这次以“当前 baseline 为唯一 baseline”为准，不再保留 `000002..000012` 原地升级兼容。
  - 确认 README / 本地测试账号文档仍然承诺默认本地 seed，因此 baseline 必须恢复这些 seed。
  - 确认完成态必须带 fresh DB replay、legacy negative-path、独立 gate review 和活跃计划文档收口证据。

## Validation

- Commands:
  - `go test ./internal/app -run 'TestImagesDeletedAtIndexInBaseline|TestActiveHostPortIndexIgnoresZeroPort|TestContestTimeContractInBaseline|TestContestStatusTransitionContractInBaseline|TestContestPausedSecondsContractInBaseline|TestRuntimeNodeContractInBaseline' -count=1`
  - `go test ./internal/app/composition -run 'TestBuildRuntimeHostExecutorProvidesReachableRuntimeInTestEnv|TestBuildContainerRuntimeModuleFailsWhenRemoteRuntimeAgentDialFails|TestBuildContainerRuntimeModuleFailsWhenLocalCheckerRunnerInitFails|TestBuildContainerRuntimeModuleProvidesDefaultRuntimeNodeSelector|TestBuildDefaultRuntimeNodeSelectorRequiresFormalMigrationOutsideTestEnv|TestBuildContainerRuntimeModuleSelectsConfiguredDefaultRuntimeNode' -count=1`
  - `go test ./tests/architecture -run 'TestAutoMigrateStaysInTestSupport' -count=1`
  - `docker exec ctf-postgres psql -U postgres -d postgres -c 'DROP DATABASE IF EXISTS ctf_submission_squash_verify;'`
  - `docker exec ctf-postgres psql -U postgres -d postgres -c 'CREATE DATABASE ctf_submission_squash_verify;'`
  - `go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3 -path ./migrations -database 'postgres://postgres:postgres123456@127.0.0.1:15432/ctf_submission_squash_verify?sslmode=disable' up`
  - `docker exec ctf-postgres psql -U postgres -d postgres -c 'CREATE DATABASE ctf_submission_squash_legacy_v12_check;'`
  - `docker exec ctf-postgres psql -U postgres -d ctf_submission_squash_legacy_v12_check -c "CREATE TABLE schema_migrations (version bigint NOT NULL PRIMARY KEY, dirty boolean NOT NULL); INSERT INTO schema_migrations(version, dirty) VALUES (12, false);"`
  - `go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3 -path ./migrations -database 'postgres://postgres:postgres123456@127.0.0.1:15432/ctf_submission_squash_legacy_v12_check?sslmode=disable' up`
  - `bash scripts/run-workflow-stage.sh pre-commit-quick`
  - `bash scripts/run-workflow-stage.sh completion-full`
- Manual checks:
  - `code/backend/schema/ctf_schema_submission.sql` contains the latest tables and columns, especially `public.runtime_nodes` and `public.instances.node_id`.
  - `docker exec ctf-postgres psql -U postgres -d ctf_submission_squash_verify -c '\dt public.runtime_nodes'`
  - `docker exec ctf-postgres psql -U postgres -d ctf_submission_squash_verify -c "SELECT column_name, data_type FROM information_schema.columns WHERE table_schema='public' AND table_name='instances' AND column_name='node_id';"`
  - `docker exec ctf-postgres psql -U postgres -d ctf_submission_squash_verify -c "SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name='contests' AND column_name IN ('status_version','paused_seconds','runtime_recovery_key','runtime_recovery_applied_seconds') ORDER BY column_name;"`
- Review focus:
  - single runtime schema owner is preserved
  - baseline remains replayable from clean state
  - deleted incremental migration references are fully removed from active docs/tests

## Execution Notes

- `pre-commit-quick` passed.
- Legacy-db negative-path verification passed: a database pinned at `schema_migrations.version=12` now fails `migrate up` with `no migration found for version 12`, which is the expected reset-required behavior for the removed chain.
- `completion-full` hit a pre-existing blocker in `internal/bootstrap/awd_defense_ssh_gateway.go`: `TestContextBackgroundOnlyAtApprovedRoots` already fails on clean `HEAD` because of existing `context.Background()` usages at lines 40, 60, 64, 99.
- The fresh-baseline replay itself succeeded on a temporary database with output `1/u init_schema`, and the replayed database contains the restored local seed (`roles_count=3`, `users_count=4`).

## Rollback Or Recovery Notes

- Git can restore `000002..000012` and pre-squash tests/docs if the new baseline proves unsuitable.
- Temporary verification databases can be dropped after validation.
