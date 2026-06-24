# CTF API Auto Migrate Implementation Plan

**Goal:** 让 `docker/docker-compose.dev.yml` 下的 `ctf-api` 容器在启动应用前先执行正式 SQL migration，并在当时的增量链下修复 `schema_migrations=11` 但 `runtime_nodes` 已被提前建表的半迁移状态。

**Architecture:** schema owner 继续收口在 `code/backend/migrations/*.sql` 和 `golang-migrate`，镜像入口只负责在 API 进程启动前调用既有 migration owner，不把 schema 变更重新塞回应用内 `AutoMigrate`。

**Tech Stack:** Dockerfile, shell entrypoint, golang-migrate, PostgreSQL migration SQL, docker compose

---

## Task Metadata

- Task Slug: `2026-06-05-ctf-api-auto-migrate`
- Started At: `2026-06-05T00:00:00Z`
- Worktree: `/home/azhi/workspace/projects/ctf`
- Branch: `task/2026-06-05-ctf-api-auto-migrate`

## Objective And Non-Goals

- Objective:
  - 给后端镜像接入 `migrate` binary 和入口脚本。
  - 在当时的增量链语境下收口 `runtime_nodes` 半迁移坏库问题；该语义现已吸收到 `000001_init_schema.up.sql` baseline 中，仅作为历史背景保留。
  - 在全容器联调路径里显式打开自动迁移，并把行为写回 README。
- Non-Goals:
  - 不把 schema owner 改回应用启动期 `AutoMigrate`。
  - 不新增单独 migration sidecar service。
  - 不改变 `code/backend/scripts/dev-run.sh --migrate` 这条本地推荐路径的 owner。

## Inputs

- Source docs:
  - `README.md`
  - `docs/README.md`
- Related architecture/contracts:
  - `code/backend/Dockerfile`
  - `code/backend/migrations/000001_init_schema.up.sql`（当前 baseline；`000012` 语义已在 2026-06-08 baseline squash 中吸收）
  - `docker/docker-compose.dev.yml`
- Related prior work:
  - `.harness/reuse-decisions/ctf-api-auto-migrate.md`
  - `code/backend/scripts/dev-run.sh`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 同时触达镜像启动链路、正式 migration、开发编排和文档说明。
  - 命中了 startup workflow 的受保护 surface，需要把 schema owner 写清楚。

## Files

- Create:
  - `code/backend/scripts/docker-entrypoint.sh`
- Modify:
  - `code/backend/Dockerfile`
  - 当时的 `000012_create_runtime_nodes.up.sql` 已在当前 baseline 中吸收，不再作为活动 runtime migration 文件存在
  - `docker/docker-compose.dev.yml`
  - `README.md`
- Review:
  - `.harness/reuse-decisions/ctf-api-auto-migrate.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-05-ctf-api-auto-migrate-implementation-plan.md`
- Test:
  - `code/backend/internal/app`
  - `code/backend/internal/module/practice/application/commands`

## 复用与 Owner 决策

- Existing patterns searched:
  - `code/backend/scripts/dev-run.sh`
  - `code/backend/migrations/000001_init_schema.up.sql`（当前 baseline；原 `000012` 语义已被吸收）
  - `docker/docker-compose.dev.yml`
- Reuse / extend / split / create-new decision:
  - 复用现有 `golang-migrate` owner，扩展镜像入口去调用它。
  - 历史上曾通过扩展当时的 `000012` migration 收口坏库状态；当前事实源已统一为 baseline，不再把已删除 migration 当成活动 owner。
- Owner boundary:
  - schema owner 仍然是 SQL migration；entrypoint 只是触发者。
  - compose 只负责为本地全容器联调打开显式开关，不承担 migration 逻辑本身。
- Why this is the narrowest safe surface:
  - 不引入新的 schema owner，也不把半迁移修复逻辑扩散到应用运行期。

## Validation

- Commands:
  - `go test ./internal/app -run TestInternalAppTestSchemaIncludesRuntimeTables -count=1`
  - `go test ./internal/module/practice/application/commands -run TestPracticeCommandDBMigratesRuntimeNodeSchema -count=1`
- Manual checks:
  - 需要时再跑 `docker compose -f docker/docker-compose.dev.yml up -d --build ctf-api`
  - 当前历史备注：`runtime_nodes` / `instances.node_id` 已在 `000001_init_schema.up.sql` baseline 中统一收口
- Review focus:
  - 正式 migration owner 是否仍然单点收口
  - entrypoint 是否只做启动前触发，而不是引入新的运行期 schema 逻辑
