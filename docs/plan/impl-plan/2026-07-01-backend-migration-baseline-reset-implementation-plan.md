<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# backend-migration-baseline-reset Implementation Plan

**Goal:** 重置后端数据库迁移 baseline，让全新 Docker 开发库只执行一个一致的 `000001_init_schema`。

**Architecture:** 迁移 owner 保持在 `code/backend/migrations/`，不改变运行时代码和 API 行为。旧增量 `000013..000020` 的结构结果被吸收到 baseline，测试从检查增量文件改为检查 baseline 契约。

**Tech Stack:** Go、PostgreSQL、go-migrate、Docker Compose

---

## Task Metadata

- Task Slug: `2026-07-01-backend-migration-baseline-reset`
- Started At: `2026-07-01T05:35:00Z`
- Worktree: `/home/azhi/workspace/projects/ctf`
- Branch: `task/2026-07-01-backend-migration-baseline-reset`
- Plan Type: `slice`

## Plan Status

- Status: `implemented`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective: 生成包含当前运行所需表、索引、默认账号和 outbox / runtime / flag secret 结构的单一 baseline，并移除已吸收的旧增量文件。
- Non-Goals: 不原地迁移生产或现有开发库；不重置用户当前 `ctf` 数据库；不修改业务查询、MCP handler 或前端代码。

## Problem Statement

- Current behavior / structure: Docker 新库执行 `000001` 后继续执行 `000013..000020`，其中 `000020` 仍假设旧 `node_id` 列存在；旧开发库又停在已删除迁移链版本，导致 API 启动失败。
- Target behavior / structure: 当前仓库的全新库只通过 `000001` 初始化到完整结构，旧链数据库仍按 entrypoint 提示重建或恢复。
- Why this task is needed now: MCP Docker 联调必须依赖可启动的 API，而当前迁移链无法稳定初始化新库。

## Inputs

- Source docs: `AGENTS.md`、`code/backend/tests/README.md`、`harness/templates/implementation-plan-skeleton.md`
- Related architecture/contracts: `code/backend/migrations/000001_init_schema.up.sql`、`code/backend/internal/app/*migration*_test.go`
- Related prior work: `docs/plan/impl-plan/2026-07-01-agent-current-challenge-mcp-implementation-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why: 修改正式数据库 baseline 和迁移护栏测试，触达受保护后端数据结构 surface。

## Files

- Create:
  - `docs/plan/impl-plan/2026-07-01-backend-migration-baseline-reset-implementation-plan.md`
- Modify:
  - `code/backend/migrations/000001_init_schema.up.sql`
  - `code/backend/internal/app/flag_secret_migration_test.go`
  - `code/backend/internal/app/migration_files_test.go`
  - `code/backend/internal/app/runtime_node_migration_test.go`
- Review:
  - `code/backend/migrations/000013_add_runtime_cluster_secret_and_flag_key_id.up.sql`
  - `code/backend/migrations/000016_create_contest_realtime_outbox.up.sql`
  - `code/backend/migrations/000018_create_platform_event_outbox.up.sql`
  - `code/backend/migrations/000019_add_runtime_node_last_seen_at.up.sql`
  - `code/backend/migrations/000020_rename_instance_runtime_node_id_and_add_awd_runtime_placements.up.sql`
- Test:
  - `go test ./internal/app -run 'Migration|Baseline|FlagSecret|RuntimeNode|EventOutbox|ContestTime|ContestStatus|ContestPaused' -count=1`
  - `go test ./internal/interfaces/mcp -count=1`
  - `go-migrate` against a fresh Docker PostgreSQL database

## 复用与 Owner 决策

- Existing patterns searched: 使用 `rg` 搜索迁移文件名、`schema_migrations`、`flag_key_id`、`last_seen_at`、outbox 表和 runtime node 迁移护栏。
- Reuse / extend / split / create-new decision: 复用现有 baseline dump 风格，重新生成 `000001`，删除已吸收的旧增量文件。
- Owner boundary: 数据结构事实由 `code/backend/migrations/000001_init_schema.up.sql` 承担，测试只检查 baseline 中必须存在的契约片段。
- Why this is the narrowest safe surface: 不改业务代码、不改 OpenAPI、不重置已有数据库；只收口初始化链路和直接相关护栏测试。

## Intake Analysis Gate

- Relevant superpowers analysis pass: 使用项目 `harness-router` 和后端规则读取流程，按 backend/migration 受保护 surface 处理。
- Why this pass fits: 任务是迁移链修复和 Docker 验证，不需要 UI 或产品设计路径。
- grill-with-docs findings: `AGENTS.md` 要求破坏性数据库操作先备份；本实现避免重置现有库，只用临时库验证。
- Plan adjustments after challenge: 将 full compose 指向临时测试库，避免破坏默认 `ctf` 库。

## Execution Slices

### Slice 1: 收口 baseline

- Goal: 生成单一 baseline 并移除旧增量迁移。
- Dependencies: Docker PostgreSQL 可用，旧 baseline 和增量 SQL 可读取。
- Files:
  - Create: 本计划文件
  - Modify: `code/backend/migrations/000001_init_schema.up.sql` 和迁移护栏测试
  - Review: 被吸收的旧增量 SQL
  - Test: 迁移相关 Go 测试和 fresh DB 迁移验证
- 步骤:
  - [x] 步骤 1：在临时库应用旧 baseline 和已吸收增量效果。
  - [x] 步骤 2：dump 为新的 `000001_init_schema.up.sql`。
  - [x] 步骤 3：删除旧增量文件并更新测试。
  - [x] 步骤 4：用 `go-migrate` 和 Docker compose 验证新库启动。
- Validation: 执行本计划 Validation Evidence 中列出的命令。
- Review focus: baseline 是否包含 MCP/API 启动所需表列，迁移目录是否只剩单 baseline。
- Done criteria: fresh DB 可迁移，API `/ready` 正常，MCP token 调用返回 JSON-RPC result。

## Impact And Compatibility

- API / DTO: none
- Data / migration: 新库 baseline 重置为当前完整结构；旧链库仍需要按 entrypoint 提示重建或恢复。
- State / cache / queue / event: baseline 包含 `contest_realtime_outbox` 与 `platform_event_outbox`。
- Runtime / config: Docker 测试使用临时数据库 override，不改变默认 compose 文件。
- Frontend route / state / UX: none
- Docs / contracts: 新增本实施计划作为本次受保护改动证据。

## Plan Review / Architecture Fit

- Target owner boundary: 数据库初始化事实集中在 `code/backend/migrations/000001_init_schema.up.sql`。
- Reuse points / landing zones: 复用现有 migration tests 和 Docker compose 本地栈。
- Known structural debt touched: 旧 sqlite 测试 fixture 缺 outbox 表导致 full `internal/app` 测试仍失败。
- How this plan avoids behavior-only convergence: 不只绕过 Docker，直接修正迁移链结构。
- Hidden second-redesign risk: 后续新增迁移时需要从新的 baseline 后继续编号，而不是恢复旧 13..20。
- Decision after review: 该切片适合以单独提交交付。

## Documentation Owner

- Current fact sources to read: `AGENTS.md`、`code/backend/tests/README.md`、migration tests。
- Fact sources to update after implementation: `docs/plan/impl-plan/2026-07-01-backend-migration-baseline-reset-implementation-plan.md`
- Plan-only notes that must not become architecture source: 临时 Docker 数据库名和本地端口 override。
- Archive condition: 提交后可按项目流程归档。

## Validation

- 计划验证范围: 迁移护栏测试、MCP handler 测试、fresh DB `go-migrate`、Docker compose 健康检查、MCP curl 调用。
- 命名 / 契约检查范围: 搜索旧迁移编号引用和 `COPY` / `\\restrict` 禁用片段。
- 完成判定: 测试通过，fresh DB 迁移后 `schema_migrations=1|false`，Docker API ready，MCP 授权调用返回 result。

## Validation Plan

- Per-slice commands:
  - `go test ./internal/app -run 'Migration|Baseline|FlagSecret|RuntimeNode|EventOutbox|ContestTime|ContestStatus|ContestPaused' -count=1`
  - `go test ./internal/interfaces/mcp -count=1`
- Integration commands:
  - `docker run --rm --network ctf-network -v "$PWD/code/backend/migrations:/work/migrations:ro" --entrypoint /app/migrate ctf-backend:local -path /work/migrations -database "postgres://postgres:postgres123456@ctf-postgres:5432/<db>?sslmode=disable" up`
  - `CTF_HOST_ROOT="$PWD" CTF_FRONTEND_PORT=5174 docker compose -f docker/docker-compose.dev.yml -f /tmp/ctf-compose-mcp-override.yml up -d --no-build --force-recreate ...`
- Manual checks:
  - `curl http://127.0.0.1:8080/ready`
  - unauth/authenticated MCP JSON-RPC curl
- Commands intentionally skipped and why: full `go test ./internal/app ./internal/interfaces/mcp -count=1` 已执行但 `internal/app` 命中既有 full-router / sqlite fixture 失败，不作为本切片完成门禁。

## Validation Evidence

- Command: `go test ./internal/app -run 'Migration|Baseline|FlagSecret|RuntimeNode|EventOutbox|ContestTime|ContestStatus|ContestPaused' -count=1 && go test ./internal/interfaces/mcp -count=1`
  - Result: passed
  - Notes: 验证迁移护栏和 MCP handler 测试。
- Command: fresh Docker PostgreSQL database + `go-migrate up`
  - Result: `1/u init_schema`, `schema_migrations=1|false`，关键表列存在。
  - Notes: 验证新 baseline 可由正式迁移工具执行。
- Command: Docker compose + MCP curl
  - Result: API/Frontend/Runtime/Gateway healthy；unauth 返回 `-32001`，Bearer token 调用返回 `has_current_challenge=false`。
  - Notes: 验证 Docker 中 MCP 可调用。

## Independent Review Handoff

- Review target: migration baseline reset and tests.
- Validation evidence summary: 见 Validation Evidence。
- Architecture / contract inputs: `code/backend/migrations/000001_init_schema.up.sql`、迁移测试。
- Known risks / review focus: full `internal/app` 现有失败未在本切片修复。
- Project-local checks to consider: `bash scripts/run-workflow-stage.sh pre-commit-quick`

## Rollback / Recovery

- Safe revert boundary: 回退本提交即可恢复旧迁移文件和旧测试。
- Data / config / runtime recovery notes: 本次未删除默认 `ctf` 数据库；已有 `/tmp/ctf-db-backups/ctf-before-mcp-docker-test-20260701-130623.dump` 可用于此前开发库恢复参考。
- Irreversible operations: none

## Residual Risks

- Risk: 现有默认 `ctf` 数据库仍是旧迁移链状态，直接用默认 compose DB 会按 entrypoint 提示重建。
- Why acceptable: 本次未获得删除默认库的明确确认，Docker 验证改用临时库避免破坏数据。
- Follow-up owner, if any: 后续如要默认库也切换，需要先确认重置或恢复策略。
