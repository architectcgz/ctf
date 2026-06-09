# Module Dependency Baseline Worktree Gate Review

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-module-dependency-baseline`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-module-dependency-baseline`
- Task slug: `2026-06-08-module-dependency-baseline`
- Plan: `docs/plan/impl-plan/2026-06-08-module-dependency-baseline-implementation-plan.md`
- Diff source: 当前 worktree 相对 `HEAD` 的未提交改动（含未跟踪文件）
- Files reviewed:
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/runtime_challenge_adapter.go`
  - `code/backend/internal/app/composition/runtime_challenge_adapter_test.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/runtime/architecture_test.go`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/module/runtime/runtime/module.go`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/plan/impl-plan/2026-06-08-module-dependency-baseline-implementation-plan.md`

## Classification Check

- Agree with `非琐碎任务`。

## Gate Verdict

- `pass`

## Findings

- 无 blocker。
- 上一轮 blocker 已收口：
  - `code/backend/internal/app/composition/runtime_module.go:45` 现在先走 `runtimeConfigOrDefault(root.Config())`，并通过 `runtimePublishedAccessHost(cfg)` 解析 published host，不再在 `cfg == nil` 时解引用 `cfg.Container`。
  - `code/backend/internal/app/composition/runtime_module_test.go:92` 新增 `TestRuntimePublishedAccessHostAllowsNilConfig`，覆盖了上一轮指出的空配置回归面。
- 上一轮 clone 防共享建议也已收口：
  - `code/backend/internal/app/composition/runtime_challenge_adapter.go:101` 现在对 `Policies` 走 `cloneCompositionTrafficPolicies`，深拷贝 `Ports` 切片。
  - `code/backend/internal/app/composition/runtime_challenge_adapter_test.go:80` 新增 `TestRuntimeChallengeTopologyAdapterClonesMutableFields`，覆盖 `Env`、`NetworkKeys`、`Resources`、`Policy Ports` 的输入变更不影响输出。

## Material Findings

- 无。

## Non-blocking Suggestions

- 无。

## Missing Validation

- 无 blocker 级验证缺口。
- 这轮独立补跑并通过的最小相关验证：
  - `go test ./internal/app/composition -run 'TestRuntimePublishedAccessHostAllowsNilConfig|TestRuntimeChallengeTopologyAdapterClonesMutableFields|TestRuntimeChallengeTopologyAdapterPreservesRuntimeFields|TestRuntimeChallengeTopologyAdapterDisablesPublishedEntryPortWithoutAccessHost|TestRuntimeChallengeSingleContainerRequestUsesSingleContainerSubnetPool' -count=1`
  - `go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1`
  - `go test ./internal/module/runtime -run 'TestRuntimeWiringDoesNotImportCrossModulePorts|TestRuntimeTestsDoNotDependOnGlobalModelOrChallengeEntity' -count=1`

## Open Questions Or Assumptions

- 假设 app composition 是允许承担 challenge/runtime 之间的 consumer-side adapter owner，这与 `01-system-architecture.md`、`07-modular-monolith-refactor.md` 中“runtime 只做模块内 wiring，app/composition 负责组合视图”的当前事实一致。
- 假设 `runtime -> challenge` baseline 的删除只要求消除 `internal/module/runtime/**` 生产代码导入；当前我用源码检索确认了这一点。

## Senior Implementation Assessment

- 这次 owner 收口方向本身是对的：
  - `runtime/runtime.Module` 不再公开 `challengeports.ImageRuntime` / `challengeports.ChallengeRuntimeProbe`，改为公开 runtime 自有 `ImageRuntimeService`、`ProvisioningService`、`CleanupService`，降低了通过字段类型把 `challenge` 依赖继续泄漏回 `runtime` 的风险。
  - challenge 适配器迁到 `internal/app/composition/` 也符合“consumer-side port adapter 由 composition 装配”的边界。
- 当前修复保持了这个方向，没有把 adapter 再塞回 `runtime`，而是在 composition 层补回空配置 guard，并把可变输入的 clone 语义显式固化到测试里，是当前最小且低风险的收口方式。

## Required Re-validation

- 无。

## Residual Risk

- 我这轮复核的范围只针对上一轮 blocker 和相邻建议，没有重新全量审视与本批次无关的所有未提交改动。
- 在该范围内，没有再发现 `nil config` 构造回归，也没有发现 challenge topology request 的明显共享可变状态问题。

## Touched Known-debt Status

- 当前事实源里没有看到与本次 touched surface 直接重叠、且要求本轮必须顺手收口的活动结构债条目。
- 本次 blocker 属于本轮迁移新引入的构造期回归，不是把旧 debt 降级成 residual risk。
