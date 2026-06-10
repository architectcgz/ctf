# AWD runtime_config 兼容层退役 Gate Review

日期：2026-06-10

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-10-awd-runtime-config-compat-retirement`
- Branch: `task/2026-06-10-awd-runtime-config-compat-retirement`
- Task: `2026-06-10-awd-runtime-config-compat-retirement`
- Plan: `docs/plan/archive/impl-plan/2026-06/2026-06-10-awd-runtime-config-compat-retirement-implementation-plan.md`
- Diff source: 当前 worktree 未提交 diff
- Files reviewed:
  - `code/backend/migrations/000015_remove_legacy_awd_runtime_config_challenge_id.up.sql`
  - `code/backend/migrations/000015_remove_legacy_awd_runtime_config_challenge_id.down.sql`
  - `code/backend/internal/app/migration_files_test.go`
  - `code/backend/internal/module/contest/application/commands/response_mappers.go`
  - `code/backend/internal/module/contest/application/queries/contest_awd_service_query.go`
  - `code/backend/internal/module/contest/application/queries/contest_awd_service_query_test.go`
  - `docs/architecture/backend/design/awd-engine-migration.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/design/backend-module-boundary-target.md`

## Classification Check

同意当前任务按非琐碎任务处理。该 diff 同时触达持久化数据迁移、AWD response/query mapper、迁移债文档和架构事实源，不能按纯文档或局部小改处理。

## Gate Verdict

Pass with review-process limitation.

未发现需要阻塞完成的 material finding。限制是：当前工具集没有可用的独立 `code-reviewer` subagent，本记录是同一会话内按 `code-reviewer` skill 做的 self-check，不能等同真正独立 gate review。

## Findings

无 material findings。

本轮 review 过程中已修正两个问题：

- `000015` 最初直接使用 `runtime_config::jsonb` 过滤和更新；由于 `runtime_config` 是 `text` 且没有 DB JSON 约束，单条历史脏值可能阻塞整个迁移。已改为 PL/pgSQL 按行解析，非法 JSON 跳过，只清理合法 JSON object 顶层 `challenge_id`。
- `docs/architecture/backend/design/awd-engine-migration.md` 仍引用 squash 前迁移号和旧的 AWD 题目字段名。已改为当前 baseline 与实际 `awd_challenge_id` 字段。

## Material Findings

无。

## Senior Implementation Assessment

当前实现把历史数据清理放在 `000015` SQL migration，mapper 只移除已经退役的过滤分支，不改变 `contest_awd_services.awd_challenge_id`、`awd_team_services.awd_challenge_id`、`awd_attack_logs.awd_challenge_id`、`awd_traffic_events.awd_challenge_id` 等当前展示/兼容字段。这个边界比继续在 mapper 里保留 sanitizer 更低风险，也避免旧存储事实继续泄漏到 response shaping owner。

迁移采用 no-op rollback 是合理的；删除 JSON key 后无法可靠重建原值，回滚文件明确说明不可逆原因。

## Required Re-validation

已执行并通过：

```bash
cd code/backend && go generate ./internal/module/contest/application/commands
cd code/backend && go test ./internal/app -run 'TestAWDRuntimeConfigChallengeIDCleanupMigration|TestImagesDeletedAtIndexInBaseline|TestActiveHostPortIndexIgnoresZeroPort|TestBaselineSeedsDefaultLocalUsers' -count=1
cd code/backend && go test ./internal/module/contest/application/queries -run TestContestAWDServiceQueryServiceListContestAWDServicesIncludesValidationState -count=1
cd code/backend && go test ./internal/module/contest/application/commands -run 'TestContestAWDServiceService(Create|Update)DoesNotPersistLegacyChallengeIDInRuntimeConfig' -count=1
cd code/backend && go test ./internal/module/contest/application/queries ./internal/module/contest/application/commands -count=1
cd code/backend && go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1
python3 scripts/check-docs-consistency.py
bash scripts/check-startup-gate.sh
bash scripts/check-workflow-governance.sh
git diff --check
```

Live database migration evidence:

```bash
docker exec ctf-postgres psql -U postgres -d ctf -v ON_ERROR_STOP=1 -P pager=off -c "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1;"
timeout 120 go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3 -path ./migrations -database 'postgres://postgres:postgres123456@127.0.0.1:15432/ctf?sslmode=disable' up
docker exec ctf-postgres psql -U postgres -d ctf -v ON_ERROR_STOP=1 -P pager=off -c "SELECT COUNT(*) AS candidate_rows FROM public.contest_awd_services WHERE runtime_config IS NOT NULL AND btrim(runtime_config) <> '' AND runtime_config LIKE '%challenge_id%';"
docker exec ctf-postgres psql -U postgres -d ctf -v ON_ERROR_STOP=1 -P pager=off -c "SELECT indexname, indexdef FROM pg_indexes WHERE schemaname = 'public' AND tablename = 'instances' AND indexname = 'idx_instances_status_updated_id';"
```

Result:

- Backup created before migration: `/home/azhi/workspace/projects/.backups/ctf-postgres/ctf-before-000014-000015-20260610-193347.sql.gz`
- Before migration: `schema_migrations.version = 13`, `dirty = false`
- Applied migrations: `14/u instances_stopping_cleanup_index`, `15/u remove_legacy_awd_runtime_config_challenge_id`
- After migration: `schema_migrations.version = 15`, `dirty = false`
- `contest_awd_services` candidate rows containing `challenge_id`: `0`
- `idx_instances_status_updated_id` exists on `public.instances(status, updated_at, id)`

## Residual Risk

- `runtime_config` 中非法 JSON 被跳过而不是强制修复。这符合本次“退役 `runtime_config.challenge_id` 兼容层”的边界，也避免数据迁移因无关历史脏值失败。

## Touched Known-debt Status

- `contest_awd_services.runtime_config.challenge_id` 兼容层：本轮已用迁移和 mapper 清理收口，不再保留为活动 backlog。
- `docs/design/backend-module-boundary-target.md`：本轮仅标记为 `Superseded`，当前事实源切到 `docs/architecture/backend/07-modular-monolith-refactor.md`；未继续把旧设计稿作为活动事实源。
