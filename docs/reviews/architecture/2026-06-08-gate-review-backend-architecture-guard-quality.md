# Backend Architecture Guard Quality Gate Review

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Task slug: `2026-06-08-backend-architecture-guard-quality`
- Plan: `docs/plan/impl-plan/2026-06-08-backend-architecture-guard-quality-implementation-plan.md`
- Diff source: 当前 worktree 相对 `HEAD b80f95fb5` 的未提交改动
- Files reviewed:
  - `code/backend/internal/testutil/archtest/archtest.go`
  - `code/backend/internal/shared/architecture_test.go`
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `code/backend/internal/module/mapper_policy_test.go`
  - `code/backend/internal/app/architecture_rules_test.go`
  - `code/backend/internal/module/contest/application/commands/awd_attack_submit_support.go`
  - `code/backend/internal/module/contest/application/commands/awd_attack_flag_rotation_support.go`
  - `scripts/check-backend-architecture.sh`
  - `scripts/lib/check-consistency/architecture.sh`
  - `docs/architecture/README.md`
  - `docs/architecture/backend/README.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/plan/impl-plan/2026-06-08-backend-architecture-guard-quality-implementation-plan.md`

## Classification Check

- Agree with `非琐碎任务`。

## Gate Verdict

- `pass with minor issues`

## Findings

### Minor

1. `docs/architecture/README.md:40`-`41` 仍把 backend quick/full 描述成“模块边界 + app/test 补充”，没有把这次新增的 `internal/shared` guard 写进入口说明。
   - 影响：后续维护者只看架构入口时，会低估 `scripts/check-backend-architecture.sh --quick/--full` 的真实覆盖面，不利于把 shared boundary 当成稳定 contract 维护。
   - 修正方向：把 quick/full 两条说明同步更新为“module + shared”，避免脚本行为和入口文案继续分叉。

2. `scripts/lib/check-consistency/architecture.sh:59` 现在只校验 backend architecture 脚本包含 `go test ./internal/shared`，但没有把这次核心收口点“module 阶段跑整个 `./internal/module` root package，而不是只跑 `TestModuleArchitectureBoundaries`”固化成一致性检查。
   - 影响：未来如果有人把 `scripts/check-backend-architecture.sh:11` 回退成单个 `-run TestModuleArchitectureBoundaries`，当前 consistency guard 仍会放行，这会削弱本轮新增的 baseline/currentness/root-context/time.Now/mapper policy coverage。
   - 修正方向：在 consistency check 中增加对 `go test ./internal/module` 全包执行语义的显式断言。

## Material Findings

- 无。当前 diff 没有发现需要阻断 completion 的 correctness / architecture blocker。

## Non-blocking Suggestions

- `code/backend/internal/shared/architecture_test.go:23`-`40` 这份 forbidden list 已覆盖本轮 plan 明确列出的 `app/bootstrap/config/model/module/platform/infrastructure/middleware/validation` 和主要 concrete infra packages，当前可以接受。
- 如果后续希望把 “shared 只能是轻量、进程无关工具” 再收紧一层，建议评估是否要把 `internal/authctx`、`internal/httpresponse`、`internal/websocket`、`internal/handler` 这类明显进程层包也纳入禁用列表，避免 boundary 只挡住部分 `internal/*` 入口。

## Missing Validation

- 无额外 blocker 级验证缺口。
- 独立 reviewer 已补跑最小相关验证：
  - `timeout 180s go test ./internal/testutil/archtest ./internal/shared ./internal/module ./internal/app -run 'TestSharedPackagesStayLightweightAndModuleAgnostic|TestModuleArchitectureBoundaries|TestModuleDependencyBaselineIsCurrent|TestDomainInternalImportExceptionsAreCurrent|TestApplicationConcreteDependencyExceptionsAreCurrent|TestCrossModulePrivateImportExceptionsAreCurrent|TestModuleRuntimeCodeDoesNotCreateRootContext|TestBackendBusinessCodeDoesNotCreateRootContext|TestTimeNowUsageExceptionsAreCurrent|TestTransactionBoundaryExceptionsAreCurrent|TestRuntimeModulesStaySmallAndWiringOnly|TestMapperWrappersFollowGlobalDelegationPolicy|TestArchitectureRulesRejectConcreteCrossModuleImports' -count=1`
  - `timeout 180s bash scripts/check-backend-architecture.sh --full`

## Open Questions Or Assumptions

- 假设 `internal/shared` 当前 contract 仍按 implementation plan 中的窄边界执行，即优先阻止业务模块、composition/root、Redis/GORM/Docker/Gin/HTTP 这类明确 owner 漂移，而不是一次性穷举所有 `internal/*` 进程层目录。
- 假设 `submitAttackContext.now` 捕获时间与 flag rotation 之间只存在请求内的短延迟，因此把 TTL 统一到同一业务时间基准不会引入可感知的 round 过期漂移；当前调用链下这一假设成立。

## Senior Implementation Assessment

- `internal/testutil/archtest` 放在 `internal/testutil/` 而不是 `internal/shared/` 是正确的 owner 选择，和仓库现有 `internal/testutil/systemapp`、`internal/testutil/runtimeadapters` 一致，没有把测试扫描逻辑误沉淀为生产共享能力。
- `scripts/check-backend-architecture.sh` 把 module 阶段切到 `go test ./internal/module` 是当前最小且正确的强化方式。这里跑的是 root package，不是递归跑整棵 `internal/module/...`，所以覆盖面扩大到了 root package 内所有 architecture/root-context/baseline guard，但成本仍维持在轻量级。
- 两个 hidden guard fix 都是在收口既有事实，而不是放宽真实问题：
  - `bootstrap/awd_defense_ssh_gateway.go` 的 `context.Background()` 的确属于独立进程入口。
  - AWD flag rotation TTL 改用 `submitAttackContext` 的业务时间后，更符合“同一业务动作使用同一 UTC 时间基准”的 guard 语义。

## Required Re-validation

- 当前 gate 无 required re-validation。
- 两处 minor issue 已在实现上下文修复：
  - `docs/architecture/README.md` 已同步 backend quick/full 覆盖范围。
  - `scripts/lib/check-consistency/architecture.sh` 已固化 `go test ./internal/module` 全包执行语义。
- 修复后已补跑：
  - `timeout 180s bash scripts/check-backend-architecture.sh --full`：通过。
  - `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`：通过。

## Residual Risk

- `internal/shared` boundary 现在主要防业务模块、composition 和几类高风险 concrete infra import；它不是“所有进程层依赖一网打尽”的穷举 guard。这个剩余空间目前属于边界收紧机会，不是现有 blocker。
- `archtest` helper 自身没有独立单测文件，但它已经被 `internal/shared`、`internal/module`、`internal/app` 的架构测试实际执行路径覆盖；当前剩余风险主要是未来有人修改 helper 时需要继续依赖这些 guard suite 提供回归保护。

## Touched Known-debt Status

- 未发现当前事实源中仍处于活动状态、且与本次 touched surface 直接重叠的结构性 debt 条目。
- 本次触达的隐藏 guard failure（root context exception、AWD business time、`time.Now` exception）都已在 touched surface 内收口，没有把同面已知 correctness 问题继续留作 residual risk。
