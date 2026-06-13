# 2026-06-13 Backend Review Context Logging Contract Round 2

- Review target:
  - Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract`
  - Branch: `task/2026-06-13-backend-context-logging-contract`
  - Commit range: `3d3abe1211e0c255c26842371c4eace8c2c5d8e5..c7c52f6126b12e0ee77fac0505abf18c4a637b63`
  - Diff basis: `3d3abe1211e0c255c26842371c4eace8c2c5d8e5..c7c52f6126b12e0ee77fac0505abf18c4a637b63`
  - Plan: `docs/plan/impl-plan/2026-06-13-backend-context-logging-contract-implementation-plan.md`
  - Files reviewed:
    - `code/backend/internal/platform/logctx/logger.go`
    - `code/backend/internal/platform/logctx/logger_test.go`
    - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
    - `code/backend/internal/module/instance/application/commands/context_logging_test.go`
    - `code/backend/tests/architecture/test_architecture_test.go`
    - `docs/plan/impl-plan/2026-06-13-backend-context-logging-contract-implementation-plan.md`
    - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
- Reviewer mode: independent gate review
- Classification check: agree with pipeline classification `非琐碎任务`
- Gate verdict: `pass with minor issues`

## Findings

### Minor

1. `code/backend/internal/platform/logctx/logger_test.go:45-61`
   - Issue: round-2 为 `logctx.Debug` 增加了 request-id 正向用例，但仍没有直接覆盖 `ctx == nil` 或 `logger == nil` 的 fallback。
   - Why it matters: `logctx.Debug` 是这轮新增 API，plan 也把 “nil context / nil logger 行为可预期” 列为 review focus。当前实现通过共享 `withContext` 复用了同一契约，但这一点主要靠代码阅读和既有非-Debug 测试推断，缺少 Debug 自身的回归护栏。
   - Suggestion: 补一条很小的单测，直接验证 `Debug(nil, nil, ...)` 或 `Debug(nil, logger, ...)` 不 panic、也不会附加脏 `request_id` 字段。

## Material Findings

None.

## Non-blocking Suggestions

1. Round-1 提到的试点内裸 `Debug` 已正确收口，可以保持当前窄范围推进。后续如果继续扩面，优先沿用 `logctx` 和同一类 source guardrail，不要重新在业务文件里手写 `request_id` field。

## Missing Validation

1. `code/backend/internal/platform/logctx/logger_test.go:45-61` 只覆盖了 `Debug` 的正向 request-id 注入，没有直接覆盖 `Debug` 的 nil fallback。

## Senior Implementation Assessment

这次 round-2 的方向是正确的。`logctx.Debug` 直接复用 [logger.go](/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract/code/backend/internal/platform/logctx/logger.go:11) 到 [logger.go](/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract/code/backend/internal/platform/logctx/logger.go:35) 的同一 `withContext` 路径，没有为 Debug 重新分叉一套 request-field 逻辑；`maintenance_service.go` 中 round-1 剩余的 4 处裸 Debug 也都已经迁移到 [maintenance_service.go](/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract/code/backend/internal/module/instance/application/commands/maintenance_service.go:142), [maintenance_service.go](/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract/code/backend/internal/module/instance/application/commands/maintenance_service.go:261), [maintenance_service.go](/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract/code/backend/internal/module/instance/application/commands/maintenance_service.go:303), [maintenance_service.go](/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract/code/backend/internal/module/instance/application/commands/maintenance_service.go:307)。pilot guardrail 仍只约束 [test_architecture_test.go](/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract/code/backend/tests/architecture/test_architecture_test.go:83) 到 [test_architecture_test.go](/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract/code/backend/tests/architecture/test_architecture_test.go:102) 里的 3 个试点文件，没有扩大成全仓规则；触达文件也没有引入对 `internal/infrastructure/logger` 的新依赖，依赖方向保持在 `internal/platform/logctx`。

## Validation Reviewed

- Inspected implementation-context evidence:
  - `timeout 240s go test ./internal/middleware ./internal/platform/requestctx ./internal/platform/logctx -count=1`
  - `timeout 240s go test ./internal/module/auth/application/commands ./internal/module/instance/application/commands ./internal/module/container_runtime/application/commands -count=1`
  - `timeout 240s go test ./tests/architecture -run 'Test(NoRawSensitiveZapFields|ContextAwareLoggingContractPilots)' -count=1`
  - `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `timeout 240s bash scripts/check-workflow-governance.sh`
- Independent reviewer reran the narrowest relevant subset:
  - `timeout 240s go test ./internal/platform/logctx ./internal/module/instance/application/commands -count=1`
  - `timeout 240s go test ./tests/architecture -run 'TestContextAwareLoggingContractPilots' -count=1`
- All rerun commands passed.

## Required Re-validation

If the minor test gap is addressed, rerun:

- `timeout 240s go test ./internal/platform/logctx -count=1`

## Residual Risk

- The pilot remains intentionally narrow and source-based. It protects the three selected files, but is not yet a repository-wide context-aware logging policy.
- The only new round-2 risk left open is test coverage depth for `logctx.Debug` nil fallback, not a correctness or ownership regression in the current implementation.

## Touched Known-debt Status

This round touches the previously reviewed pilot logging surface and closes the specific round-1 follow-up on raw `Debug` calls in `instance/application/commands/maintenance_service.go`. I did not find a current fact-source entry requiring broader structural debt closure in the touched surface for this round.
