# 2026-06-20 Backend Review Goroutine Panic Owner Round 1

## Review 对象

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-20-backend-goroutine-panic-owner`
- Branch: `task/2026-06-20-backend-goroutine-panic-owner`
- Commit: `edf54ad5e45d51e8754c1d294aeac3b8c12335a3`
- Task / Plan: `2026-06-20-backend-goroutine-panic-owner`, `docs/plan/archive/impl-plan/2026-06/2026-06-20-backend-goroutine-panic-owner-implementation-plan.md`
- Reviewer mode: same-context commit-bound review
- Diff basis: `git show edf54ad5e45d51e8754c1d294aeac3b8c12335a3`

## 结论

Gate verdict: pass with process limitation.

未发现 blocker、major finding 或 material finding。限制是：本轮 review 绑定了明确 commit，但当前工具上下文没有独立 reviewer agent，因此它不能冒充独立 gate review；它是提交绑定的 same-context review 证据。

## Findings

None.

## Material Findings

None.

## Senior Implementation Assessment

该提交删除 `internal/shared/safego`，没有引入替代性共享 goroutine helper。root 级后台任务在 owner 内记录 panic 后重新抛出，practice async task 与 runtime cleaner 维持显式 goroutine / WaitGroup 生命周期，符合“panic recovery 是 owner 失败语义，不是启动语法糖”的规则。

架构测试从旧的“试点必须导入 safego”改成“禁止 `internal/shared/safego`、对应 import 与 `safego.Go(` 调用”，更贴合当前边界。`.agents/skills/ctf-backend-patterns/SKILL.md` 已包含 root re-panic、wait-only goroutine、业务失败状态 recover 和禁止共享 helper 的代码范例。

## 必须重跑的验证

提交前后已执行并通过：

- `go test ./internal/app/composition ./internal/module/practice/application/commands ./internal/module/instance/infrastructure ./tests/architecture -count=1`
- `go test -race ./internal/app/composition -run 'TestBackgroundJobStartUsesProvidedContext|TestNewLoopBackgroundJobRejectsNilContext' -count=1`
- `go test -race ./internal/module/practice/application/commands -run 'TestPracticeServiceRunAsyncTaskReturnsWhenClosed|TestPracticeServiceCloseCancelsAsyncScoreUpdate' -count=1`
- `go test -race ./internal/module/instance/infrastructure -run 'TestCleanerStopCancelsRunningTask|TestCleanerStopsRunningTaskWhenCleanupLockIsLost|TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation' -count=1`
- `rg -n '"ctf-platform/internal/shared/safego"|safego\.Go' code/backend/internal` returned no matches.
- `bash scripts/check-shared-skills.sh`
- `bash scripts/run-workflow-stage.sh completion-full`
- `bash scripts/check-workflow-governance.sh`
- `bash scripts/run-workflow-stage.sh pre-commit-quick`

## 残余风险

- root background job panic 现在会在记录后重新抛出并导致进程失败；这是本任务明确选择的 owner 语义。如果后续需要自动重启、降级或 supervisor 策略，应另开 root lifecycle 设计任务。
- 当前 review 不是独立 reviewer gate。若严格要求独立 gate，应在可用独立 reviewer agent 的上下文中对 `edf54ad5e45d51e8754c1d294aeac3b8c12335a3` 再审一轮。

## Touched Known-debt Status

本提交直接移除了已登记为试点债务的共享 `SafeGo` helper，并把可复用规则沉淀到 CTF 后端 harness skill 与 archived feedback。未发现新的同面结构债。
