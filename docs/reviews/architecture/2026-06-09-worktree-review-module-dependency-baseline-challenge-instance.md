# Module Dependency Baseline Challenge-Instance Worktree Review

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-module-dependency-baseline`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-module-dependency-baseline`
- Task slug: `2026-06-08-module-dependency-baseline`
- Plan: `docs/plan/archive/impl-plan/2026-06/2026-06-08-module-dependency-baseline-implementation-plan.md`
- Diff source: 当前 worktree 相对 `HEAD` 的未提交改动
- Files reviewed:
  - `code/backend/internal/module/challenge/infrastructure/repository.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/challenge/infrastructure/repository_test.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_service.go`
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/module/instance/contracts/persistence.go`
  - `code/backend/internal/module/instance/entity/instance.go`
  - `code/backend/internal/infrastructure/postgres/postgres.go`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`

## Classification Check

- Agree with `非琐碎任务`。

## Gate Verdict

- `pass`

## Findings

- 无 blocker。
- `code/backend/internal/module/challenge/infrastructure/repository.go:206`
  `HasRunningInstances` 从 `Model(&instancecontracts.Instance{})` 改为 `Table("instances")` 后，当前仓库内行为保持等价：
  - `instance/contracts.Instance` 是 `instance/entity.Instance` 的 type alias。
  - `instance/entity.Instance` 没有 `DeletedAt` / soft-delete scope，也没有自定义 `TableName()`。
  - 当前 GORM 初始化没有自定义 `NamingStrategy` 或 `TablePrefix`。
  - 查询条件 `challenge_id` 与 `status IN ("creating","running")` 未变化。
- `code/backend/internal/module/architecture_baseline_test.go:9`
  baseline 删除 `challenge -> instance` 有真实 import 消失支撑。独立复核时对 `internal/module/challenge` 生产代码执行了更宽的 `rg 'ctf-platform/internal/module/instance(/|\"|$)' --glob '!**/*_test.go' --glob '!**/testsupport/**'`，结果为空。

## Material Findings

- 无。

## Non-blocking Suggestions

- 无。

## Missing Validation

- 无 blocker 级验证缺口。
- 实现上下文已提供并通过的验证：
  - `go test ./internal/module/challenge/infrastructure ./internal/module/challenge/application/commands -run 'TestServiceDeleteChallengeWithRunningInstances|Test.*HasRunningInstances|Test.*DeleteChallenge' -count=1`
  - `go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `rg -n 'ctf-platform/internal/module/instance' internal/module/challenge --glob '!**/*_test.go' --glob '!**/testsupport/**' -S`
  - `bash scripts/check-backend-architecture.sh --full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- 本轮独立补跑并通过的最小相关验证：
  - `go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `go test ./internal/module/challenge/infrastructure -run TestRepositoryHasRunningInstances -count=1`

## Open Questions Or Assumptions

- 假设本任务当前目标只是消除 `challenge -> instance` 的真实 import，而不是同时把“运行中实例判断”抽成新的 challenge-owned port 或 shared contract；这一点与 plan 中“先做 cheap leak batch”的阶段目标一致。
- 假设当前 PostgreSQL / SQLite 都继续使用默认 `instances` 表名；如果未来引入统一 `NamingStrategy`、schema prefix 或 `TableName()`，这里需要同步回看。

## Senior Implementation Assessment

- 这次收口方式是当前最小且可审阅的实现：
  - 没有放松 `TestModuleDependencyBaselineIsCurrent`。
  - 没有把 `instance` 的业务契约重新搬到 `challenge`。
  - 只把 `challenge` 对 `instance` 的耦合从 Go import 降到稳定表名读取，满足了本批次“真实 import 消失后再删 baseline”的要求。
- 在当前仓库约束下，这比为了一个 existence check 再引入新的跨模块 capability port 更低成本，也更符合本批次的收口粒度。

## Required Re-validation

- 无。

## Residual Risk

- `Table("instances")` 依赖当前默认表名约定，后续若引入命名策略、表前缀或 entity 自定义 `TableName()`，这处不会自动跟随；这不是当前 diff 的行为回归，但属于后续架构演进时需要记得回看的点。
- 补测后，`code/backend/internal/module/challenge/infrastructure/repository_test.go:118` 已显式覆盖 `creating => true`、`running => true`、`stopped => false`。当前残余风险不再是状态集合覆盖不足，而只剩表名约定这一基础设施前提。

## Touched Known-debt Status

- 本次 touched surface 命中了已知结构债 `moduleDependencyBaseline`。
- 在 `challenge -> instance` 这一条边上，真实生产 import 已移除，baseline 删除与源码现状一致，这条 debt 在当前 touched surface 内已收口，没有被降级成 residual risk。
