# 2026-06-20 Backend Review Goroutine Panic Owner Draft

## Review 对象

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-20-backend-goroutine-panic-owner`
- Branch: `task/2026-06-20-backend-goroutine-panic-owner`
- Commit: 未提交工作区 diff
- Task / Plan: `2026-06-20-backend-goroutine-panic-owner`, `docs/plan/impl-plan/2026-06-20-backend-goroutine-panic-owner-implementation-plan.md`
- Reviewer mode: draft review on current worktree diff
- Diff basis: `git diff` + untracked files listed by `git status --short --branch`
- Files reviewed:
  - `.agents/skills/ctf-backend-patterns/SKILL.md`
  - `feedback/2026-06-20-goroutine-panic-owner-boundary.md`
  - `code/backend/internal/app/composition/root.go`
  - `code/backend/internal/app/composition/root_background_job_test.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner.go`
  - `code/backend/tests/architecture/test_architecture_test.go`

## 结论

Current diff 预检结论：pass for current diff.

未发现 blocker 或 material finding。该记录只审查未提交工作区 diff，按项目文档规范不能单独充当正式 gate review；如果后续提交本任务，需要对提交后的 commit 或 commit range 再补正式 round。

## Findings

None.

## Material Findings

None.

## Senior Implementation Assessment

本次改动没有用新的共享 helper 替换 `safego`，而是把 goroutine 启动、取消、等待和 panic 后果收回到各自 owner。`root.go` 在 root background job 内记录 `background_job_panicked` 后重新 panic，符合关键后台能力不能静默死亡的规则；`WaitGroup.Wait()` 辅助 goroutine 回到裸等待；practice async task 和 runtime cleaner 不再依赖共享 recover 默认值，保留现有取消和等待语义。

规则沉淀落在 `.agents/skills/ctf-backend-patterns/SKILL.md` 的 Concurrency & Durable State 段，并有 `feedback/2026-06-20-goroutine-panic-owner-boundary.md` 保存背景。架构测试改为禁止 `internal/shared/safego` 包、导入和 `safego.Go(` 调用，方向上比旧的“必须导入 helper” guard 更贴近当前 owner 边界。

## 必须重跑的验证

当前 diff 下已重跑并通过：

- `go test ./tests/architecture -run TestNoSharedSafeGoDefault -count=1`
- `go test ./internal/app/composition -run 'TestBackgroundJobStartUsesProvidedContext|TestNewLoopBackgroundJobRejectsNilContext' -count=1`
- `go test ./internal/module/practice/application/commands -run 'TestPracticeServiceRunAsyncTaskReturnsWhenClosed|TestPracticeServiceCloseCancelsAsyncScoreUpdate' -count=1`
- `go test ./internal/module/instance/infrastructure -run 'TestCleanerStopCancelsRunningTask|TestCleanerStopsRunningTaskWhenCleanupLockIsLost|TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation' -count=1`
- `go test ./internal/app/composition ./internal/module/practice/application/commands ./internal/module/instance/infrastructure ./tests/architecture -count=1`
- `go test -race ./internal/app/composition -run 'TestBackgroundJobStartUsesProvidedContext|TestNewLoopBackgroundJobRejectsNilContext' -count=1`
- `go test -race ./internal/module/practice/application/commands -run 'TestPracticeServiceRunAsyncTaskReturnsWhenClosed|TestPracticeServiceCloseCancelsAsyncScoreUpdate' -count=1`
- `go test -race ./internal/module/instance/infrastructure -run 'TestCleanerStopCancelsRunningTask|TestCleanerStopsRunningTaskWhenCleanupLockIsLost|TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation' -count=1`
- `rg -n '"ctf-platform/internal/shared/safego"|safego\.Go' code/backend/internal` returned no matches.
- `bash scripts/check-shared-skills.sh`
- `bash scripts/check-startup-gate.sh`
- `bash scripts/run-workflow-stage.sh completion-full`
- `bash scripts/check-workflow-governance.sh`

如果后续在提交前继续改动本任务文件，应重跑受影响的最小集合；正式 gate review 仍需绑定提交后的 commit 或 commit range。

## 残余风险

- 正式 gate review 尚未绑定 commit/range；该 draft 只能作为实现中预检证据。
- root background job panic 现在会在记录后重新抛出并导致进程失败，这是规则要求的显式失败语义；如果后续需要 supervisor / restart 策略，应另开 root lifecycle 任务。
- practice async task 和 runtime cleaner 当前选择是不共享 recover 默认值。若未来要把某类 panic 转成业务失败或单轮失败重试，应在对应 owner 内加本地 recover 和测试，不回到共享 SafeGo。

## Touched Known-debt Status

本次直接移除的 `internal/shared/safego` 试点就是 touched known debt。当前 diff 已删除 helper、删除旧 helper 导入 guard，并用项目 skill + architecture test 固化“不默认共享 SafeGo”的 owner 规则。
