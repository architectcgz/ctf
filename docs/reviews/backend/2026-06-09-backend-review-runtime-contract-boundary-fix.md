# Runtime Contract Boundary Fix Backend Review

- Review target:
  - Repository: `/home/azhi/workspace/projects/ctf`
  - Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-runtime-contract-boundary-fix`
  - Task slug: `2026-06-09-runtime-contract-boundary-fix`
  - Diff source: 当前 task worktree 未提交改动
  - Files reviewed:
    - `code/backend/internal/app/composition/practice_module.go`
    - `code/backend/internal/module/practice/runtime/module.go`
    - `code/backend/internal/module/practice/infrastructure/repository.go`
    - `code/backend/internal/module/architecture_test.go`
    - `code/backend/internal/module/runtime/contracts/portreservation/owner.go`
    - `code/backend/internal/testutil/systemapp/practice_flow.go`
    - `code/backend/internal/module/practice/application/commands/runtime_port_owner_test.go`
    - `code/backend/internal/module/practice/application/commands/runtime_port_owner_external_test.go`
    - 相关 `practice/application/commands/*_test.go`
    - `docs/plan/impl-plan/2026-06-09-runtime-contract-boundary-fix-implementation-plan.md`
  - Related docs checked:
    - `docs/architecture/backend/07-modular-monolith-refactor.md`
    - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`

- Classification check:
  - Agree with `非琐碎任务`.

- Gate verdict:
  - `pass`

## Findings

- 无 blocker finding。

## Non-blocking suggestions

1. `code/backend/internal/module/runtime/contracts/portreservation/owner.go:1`
   - 当前只保留一个空 package 以避免未确认删除，这不会重新引入 concrete 依赖。
   - 但这个目录名仍然像“有效 contract namespace”，后续读代码的人容易误判这里还承载对外 contract。
   - 建议在允许删除文件的任务里直接删掉该目录，或者至少补一个简短注释说明它是待清理兼容壳。

## Missing validation

1. 目前没有一条窄测试专门锁住 `practice` 生产装配必须传入 `RuntimePortOwnerFor`。
   - 这次我独立补跑了 `TestPracticeFlow_PublishedChallengeLifecycleAndAccess`，足以证明当前生产 wiring 可用。
   - 但 `RuntimePortOwnerFor` 仍是可选字段，未来如果有人在 `composition/practice_module.go` 误删这条接线，最先兜底的仍是集成链路，不是更窄的 wiring guard。
   - 建议后续补一条 source-level 或 focused wiring test，把 `RuntimePortOwnerFor:` 这个 marker 锁住。

## Open questions or assumptions

1. 假设当前非测试生产入口只有 `internal/app/composition/practice_module.go` 会构造 `practice` 模块。
   - 我用 `rg` 复核了非测试代码，未发现其它 `practiceruntime.Build(...)` 或 `practiceinfra.NewRepository(db)` 的生产调用点。
   - 如果后续新增手工装配入口，必须同步传入 `RuntimePortOwnerFor` 或直接使用 `NewRepositoryWithRuntimePortOwner(...)`。

## Material findings

- 无。

## Senior implementation assessment

- 这次 owner 收口方式是当前最小且正确的实现：
  - concrete `runtime` repository 的创建被上移回 app composition / explicit wiring；
  - `practice/infrastructure.Repository` 不再通过 `contracts` 反向创建 runtime concrete；
  - `WithDB(tx)` 会基于同一个 `*gorm.DB` 重新构造 `runtimeports.PortReservationOwner`，因此 start/restart 链路里的端口保留和实例状态更新仍在同一 DB/tx handle 上完成。
- 我重点核对了下面三处：
  - `code/backend/internal/app/composition/practice_module.go:12-37`
  - `code/backend/internal/module/practice/runtime/module.go:55-57,143-145`
  - `code/backend/internal/module/practice/infrastructure/repository.go:40-66,419-465,555-587`
- 结论是：contracts/ports/domain 的 concrete 泄漏已经被消除，生产 wiring 也没有把端口保留从事务 handle 上拆出去。

## Required re-validation

- 无返修要求。
- 实现上下文已提供的验证证据我已审阅。
- 独立 reviewer 额外执行并通过：
  - `go test ./internal/module -run TestBoundaryPackagesDoNotDependOnOuterLayers -count=1`
  - `go test ./internal/module/practice/infrastructure -count=1`
  - `go test ./internal/app -run 'TestPracticeModuleUsesTypedPortsDeps|TestPracticeModuleUsesTypedCrossModuleDeps' -count=1`
  - `go test ./internal/app -run 'TestPracticeFlow_PublishedChallengeLifecycleAndAccess' -count=1`

## Residual risk

- `practiceinfra.NewRepository(db)` 现在是“部分能力可用、端口相关能力需显式接线”的构造器。当前生产代码没有误用它，但这个风险仍依赖代码评审和后续 guardrail，而不是类型系统。
- 空的 `runtime/contracts/portreservation` 兼容壳已在 review 后按用户确认删除。

## Touched known-debt status

- 本次 touched surface 命中了架构文档里明确要求收口的边界问题：`contracts` / `ports` / `domain` 不应反向依赖 outer-layer concrete，`runtime` concrete 应由 composition/runtime wiring owner 装配。
- 当前 diff 已把这个 touched debt 在本轮范围内收口；未发现同 surface 上仍残留 blocker 级结构债。

## Post-review follow-up

- Review 后已按用户确认删除空的 `runtime/contracts/portreservation` 兼容壳。
- Review 中建议的 focused wiring guard 已补入 `TestPracticeModuleWiresRuntimePortOwnerFromCompositionRoot`。
