# Container Runtime Tail Migration Architecture Review

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-10-container-runtime-tail-migration`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-10-container-runtime-tail-migration`
- Branch: `task/2026-06-10-container-runtime-tail-migration`
- Task slug: `2026-06-10-container-runtime-tail-migration`
- Plan: `docs/plan/archive/impl-plan/2026-06/2026-06-10-container-runtime-tail-migration-implementation-plan.md`
- Diff source: 当前 worktree 相对 `HEAD` 的未提交改动（含未跟踪 plan/review 文档）
- Files reviewed:
  - `code/backend/internal/module/container_runtime/**`
  - `code/backend/internal/module/runtime/**`
  - `code/backend/internal/app/composition/**`
  - `code/backend/internal/app/architecture_rules_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/{challenge,contest,practice,instance}/**`
  - `code/backend/internal/testutil/**`
  - `code/backend/tests/system/http/**`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-10-container-runtime-tail-migration-implementation-plan.md`

## Classification Check

- Agree with `非琐碎任务 / structural backend migration`。
- 理由：这次变更迁移 container capability 的 owner、删除旧 capability import path、调整 app composition 装配、更新 module dependency baseline / architecture guard / 架构事实文档，并触达当前 `runtime` / `container_runtime` 边界技术债。

## Gate Verdict

- Same-context review verdict: `pass with no material findings`
- Independent gate status: not satisfied in this tool context. 当前会话没有可用的独立 reviewer subagent 工具；本文件只能作为 implementation context 的架构自审证据，不能伪装成真正独立的 `code-workflow` review gate。

## Findings

- No blocker findings found in the same-context architecture self-check.

## Material Findings

- None found in the same-context review.

## Senior Implementation Assessment

- 迁移方向与用户要求一致：旧 `runtime/application`、`runtime/application/commands`、`runtime/ports`、`runtime/domain`、`runtime/agentcontracts`、`runtime/infrastructure/{agentclient,agentserver}` 能力实现路径已迁到 `container_runtime`，旧 `runtime/runtime` production wrapper 也已删除，没有再为旧能力路径保留 compatibility shim。
- 留在 `runtime` 的内容主要是 persistence/state owner：`runtime/contracts/persistence.go`、AWD workspace / proxy traffic 相关 ports、runtime entity 与 GORM / Redis / state repositories。这个边界避免把 instance、practice、contest 的混合业务状态错误下沉到 `container_runtime`。
- `container_runtime/runtime.Module` 只组合显式能力端口和服务；app composition 负责把 runtime persistence/state repository 注入到 capability services，避免 module builder 内部直接 new 宽 runtime repository。
- 架构 guard 已覆盖两个关键约束：`container_runtime` 必须拥有 capability packages；`container_runtime/runtime` 不能再依赖旧 `runtime/application|ports`。`runtime/architecture_test.go` 也明确要求已退场的 `runtime/runtime` production package 不回流。

## Validation Evidence

Implementation context already executed these checks successfully:

```bash
cd code/backend && timeout 180s go test ./internal/module/container_runtime/... -count=1
cd code/backend && timeout 180s go test ./internal/module/runtime/... -count=1
cd code/backend && timeout 180s go test ./internal/app/composition -count=1
cd code/backend && timeout 180s go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestCrossModulePrivateImportExceptionsAreCurrent|TestModuleArchitectureBoundaries|TestRuntimeHostExecutorUsageIsRestricted' -count=1
cd code/backend && timeout 180s go test ./internal/app -run 'TestArchitectureRules.*|TestRouterCompositionStructure|TestRuntime.*|TestBuildContainerRuntimeModule.*' -count=1
rg -n "ctf-platform/internal/module/runtime/(application|ports|agentcontracts|domain|infrastructure/agentclient|infrastructure/agentserver)" code/backend/internal code/backend/tests -g '*.go' -g '!*_test.go'
cd code/backend && timeout 300s go test ./internal/module/challenge/... ./internal/module/contest/... ./internal/module/practice/... ./internal/module/instance/... -count=1
timeout 900s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full
python3 scripts/check-docs-consistency.py
git diff --check
```

Recorded result: all Go / workflow / docs / diff checks above passed; the `rg` production old-capability-import scan returned no matches.

## Required Re-validation

- After this review document and plan wording updates, rerun:

```bash
timeout 1200s bash scripts/check-workflow-complete.sh
timeout 120s git diff --check
```

- If workflow completion detects stale checklist or review governance issues, fix those and rerun the same gate.

## Residual Risk

- This review was produced in the implementation context, so it is weaker than a separate `code-reviewer` agent review.
- The remaining active debt is not a capability migration gap; it is the mixed runtime persistence/state owner split documented in `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`.
- No database schema or runtime-agent wire contract changes were intended in this batch; verification focused on compile-time architecture boundaries and existing module behavior tests rather than live Docker end-to-end execution.

## Touched Known-debt Status

- Touched debt: `docs/todos/2026-05-17-project-tech-debt-from-migrations.md` runtime / container_runtime boundary item.
- Status after this task: capability contracts, ports, application services, domain helpers, Docker host adapter, runtime-agent protocol / bridge, and host-executor adapters are owned by `container_runtime`; old capability implementation paths are retired instead of kept as compatibility shims.
- Remaining debt: runtime persistence/state records whose final owner is mixed across instance / AWD / proxy traffic flows.
