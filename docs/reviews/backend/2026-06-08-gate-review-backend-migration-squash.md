# Backend Migration Squash Gate Review

- Review target:
  - Repository: `/home/azhi/workspace/projects/ctf`
  - Branch: `main`
  - Diff source: 当前 working tree 未提交 diff
  - Excluded from scope: `feedback/2026-06-07-retry-backoff-should-not-depend-on-logger-presence.md`
  - Files reviewed:
    - `README.md`
    - `code/backend/cmd/import-challenge-packs/main.go`
    - `code/backend/internal/app/composition/runtime_module.go`
    - `code/backend/internal/app/composition/runtime_module_test.go`
    - `code/backend/internal/app/migration_files_test.go`
    - `code/backend/internal/app/contest_status_transition_migration_test.go`
    - `code/backend/internal/app/contest_paused_seconds_migration_test.go`
    - `code/backend/internal/app/runtime_node_migration_test.go`
    - `code/backend/migrations/000001_init_schema.up.sql`
    - `code/backend/migrations/000001_init_schema.down.sql`
    - `code/backend/schema/ctf_schema_submission.sql`
    - `code/backend/scripts/dev-run.sh`
    - `code/backend/scripts/docker-entrypoint.sh`
    - `code/backend/tests/architecture/test_architecture_test.go`
    - `docs/architecture/backend/02-database-design.md`
    - `docs/architecture/backend/design/awd-engine-migration.md`
    - `docs/plan/impl-plan/2026-06-05-ctf-api-auto-migrate-implementation-plan.md`
    - `docs/plan/impl-plan/2026-06-08-backend-migration-squash-implementation-plan.md`
    - `docs/requirements/local-dev-test-credentials.md`
    - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`

- Classification check:
  - Agree with `非琐碎任务` classification.

- Gate verdict:
  - `pass`

- Re-review update:
  - `2026-06-08` 复核已确认原 blocker 解除。
  - `docs/plan/impl-plan/2026-06-05-ctf-api-auto-migrate-implementation-plan.md` 已把 `000012` 改成历史背景 / 已吸收到 baseline 的备注，不再作为活动 owner 或复用目标。
  - `docs/plan/impl-plan/2026-06-08-backend-migration-squash-implementation-plan.md` 已把 goal 改成“把旧链直到 `000012` 才补齐的问题吸收到 baseline”，不再把 `000012` 当成当前 replay 目标。
  - 复核 `rg` 结果后，剩余命中均属于现行 reset/recreate 提示、当前 task plan 的删除清单 / 兼容性 / rollback 说明，或明确声明“已吸收到 baseline、不再作为活动 runtime migration 文件存在”的历史备注。

## Findings

- 本次复核未发现剩余 blocker。

## Material findings

- 无。

## Senior implementation assessment

- 代码侧的 owner 收口方向是对的：`000001_init_schema.up.sql` 已吸收目标 schema，`dev-run.sh` / `docker-entrypoint.sh` 也明确把旧链升级改成 reset/recreate 路径，`runtime_module.go` 与 `cmd/import-challenge-packs/main.go` 中的隐式 `AutoMigrate` 兜底也被拿掉，并由架构测试护栏限制回流。
- 当前阻塞点不是代码 correctness，而是活动计划文档没有一起收口到同一事实源，导致“唯一 baseline”契约仍有一处文档 owner 漂移。

## Required re-validation

- 复核已基于上述 `rg` 结果完成；当前不再有额外必需 re-validation。

## Residual risk

- 独立复核了最直接相关的两组测试：
  - `go test ./internal/app -run 'TestImagesDeletedAtIndexInBaseline|TestActiveHostPortIndexIgnoresZeroPort|TestBaselineSeedsDefaultLocalUsers|TestContestTimeContractInBaseline|TestContestStatusTransitionContractInBaseline|TestContestPausedSecondsContractInBaseline|TestRuntimeNodeContractInBaseline' -count=1` -> passed
  - `go test ./tests/architecture -run 'TestAutoMigrateStaysInTestSupport' -count=1` -> passed
- 没有由 reviewer 重新跑 fresh DB replay 或 legacy negative-path；这部分仍依赖用户提供的验证证据。
- `cmd/import-challenge-packs/main.go` 去掉 `AutoMigrate` 后未见独立 smoke test，但当前 review 没发现仓库内还有把它当作“空库自举工具”的活动文档契约。

## Touched known-debt status

- 本次 touched surface 未命中需要在同面继续拆分的已知结构债。
- 原阻塞点已在活动计划文档中收口，不再构成 gate blocker。
