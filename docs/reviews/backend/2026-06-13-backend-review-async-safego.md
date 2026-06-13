# 2026-06-13 Backend Review Async SafeGo

- Review target:
  - Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-async-safego`
  - Branch: `task/2026-06-13-backend-async-safego`
  - Commit range: `ef83105c79728cceabc4e58fe150eca3c28c5e45..working-tree`
  - Diff basis: `ef83105c79728cceabc4e58fe150eca3c28c5e45..working-tree`
  - Plan: `docs/plan/impl-plan/2026-06-13-backend-async-safego-implementation-plan.md`
  - Files reviewed:
    - `code/backend/internal/app/composition/background_job_loop.go`
    - `code/backend/internal/app/composition/root.go`
    - `code/backend/internal/shared/safego/safego.go`
    - `code/backend/internal/shared/safego/safego_test.go`
    - `code/backend/internal/app/composition/root_safego_test.go`
    - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
    - `code/backend/internal/module/practice/application/commands/service_lifecycle_safego_test.go`
    - `code/backend/internal/module/instance/infrastructure/cleaner.go`
    - `code/backend/internal/module/instance/infrastructure/cleaner_safego_test.go`
    - `code/backend/tests/architecture/test_architecture_test.go`
    - `docs/plan/impl-plan/2026-06-13-backend-async-safego-implementation-plan.md`
    - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
- Reviewer mode: independent gate review
- Classification check: agree with pipeline classification `非琐碎任务`
- Gate verdict: `pass`

## Findings

None.

## Material Findings

None.

## Non-blocking Suggestions

1. 后续如果继续扩面到更多后台任务，优先保持 `SafeGo` 只负责 recover 和同步，不要把 timeout、retry、metrics 或任务调度语义继续堆进这个 helper。

## Missing Validation

None.

## Senior Implementation Assessment

这轮实现把共享 recover owner 落在 `internal/shared/safego`，同时守住了 shared package 不能反向依赖 `internal/platform/*` 的边界；panic log 只记录 `panic`、`stack` 和 `task_name`，request-scoped 字段继续留给调用方 logger owner 决定。`root.go` 试点只把显式带 logger 的平台事件后台 job 切到 SafeGo，默认两参 `NewLoopBackgroundJob` 被保留在独立文件里，避免未纳入 slice4 的其他 background job 被悄悄改成 recover 且无日志。`practice` 的 async task 和 `instance cleaner` 的 run/stop wait 路径都统一走 SafeGo，并通过源码 guardrail 锁住三处试点文件不再回退到裸 `go func()`。

## Validation Reviewed

- Inspected implementation-context evidence:
  - `go test ./internal/shared/safego -count=1`
  - `go test ./internal/app/composition -run 'TestNewLoopBackgroundJobLogsPanicThroughSafeGo|TestBackgroundJobStartUsesProvidedContext|TestNewLoopBackgroundJobRejectsNilContext' -count=1`
  - `go test ./internal/module/practice/application/commands -run 'TestRunAsyncTaskRequiresSafeGoForPanicRecovery|TestPracticeServiceRunAsyncTaskReturnsWhenClosed|TestPracticeServiceCloseCancelsAsyncScoreUpdate' -count=1`
  - `go test ./internal/module/instance/infrastructure -run 'TestCleanerStartRunOnceRequiresSafeGoRecovery|TestCleanerStopCancelsRunningTask|TestCleanerStopsRunningTaskWhenCleanupLockIsLost|TestCleanerStopReleasesCleanupLockAfterBaseContextCancellation' -count=1`
  - `go test ./tests/architecture -run 'TestSafeGoContractPilots' -count=1`
  - `go test ./internal/shared/safego ./internal/app/composition ./internal/module/practice/application/commands ./internal/module/instance/infrastructure ./tests/architecture -count=1`
  - `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `timeout 240s bash scripts/check-workflow-governance.sh`
- Independent reviewer reran the narrowest relevant subset:
  - 复核 diff 与验证记录；未发现需要额外重跑的 blocker。

## Required Re-validation

None.

## Residual Risk

- 本 slice 仍然只覆盖三个高风险后台入口；仓库内其他裸 goroutine 仍存在，后续需要继续按 slice 推进。
- `SafeGo` 对 `nil ctx` 只提供 inert context 兜底，目的是避免 helper 自己伪造 root context；这条契约已经有单测覆盖，但后续扩面时仍需保持同样边界。

## Touched Known-debt Status

本次触达的是 task group 已明确登记的后台 goroutine recover 技术债。共享 helper、三处试点迁移和源码 guardrail 已在 touched surface 内收口，没有留下新的同面债务。
