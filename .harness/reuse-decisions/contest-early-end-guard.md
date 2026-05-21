# Reuse Decision

## Change type

service / contract / test

## Existing code searched

- `code/backend/internal/module/contest/application/commands/contest_update_commands.go`
- `code/backend/internal/module/contest/application/commands/contest_update_support.go`
- `code/backend/internal/module/contest/application/commands/contest_service_test.go`
- `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- `code/backend/internal/module/contest/infrastructure/ended_contest_runtime_cleaner.go`

## Similar implementations found

- `code/backend/internal/module/contest/application/commands/awd_readiness_gate.go`
- `code/backend/internal/module/contest/application/commands/contest_update_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`

## Decision

extend_existing

## Reason

这次不是要新增新的 contest command，也不是把比赛结束清理链路拆成另一套安全网。问题入口已经很清楚，在现有 `ContestService.UpdateContest` 手动状态更新路径上。

最小正确方案是复用现有的：

- `UpdateContestInput` 里的 `ForceOverride / OverrideReason`
- `validateContestUpdateRequest(...)` 的状态校验入口
- `contest_service_test.go` 现有的手动更新回归测试结构

在原有命令里补“运行中提前结束必须显式 override”的约束，而不是新增 handler、repo 或旁路 service。这样能把风险收口在唯一写入口，不复制状态判断，也不改自动状态机。

## Files to modify

- `.harness/reuse-decisions/contest-early-end-guard.md`
- `docs/plan/impl-plan/2026-05-21-contest-early-end-guard-implementation-plan.md`
- `code/backend/internal/module/contest/application/commands/contest_update_support.go`
- `code/backend/internal/module/contest/application/commands/contest_service_test.go`
- `code/backend/internal/module/contest/contracts/errors.go`

## After implementation

- 运行中的比赛不会再因为一次普通手动更新被直接切成 `ended`
- 必须显式 override 才能提前结束比赛
- 回归测试可以覆盖这条风险，不再只能靠赛后排查数据库定位
