# Container Runtime Module Boundary Independent Gate Review

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-container-runtime-module-boundary`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-container-runtime-module-boundary`
- Task slug: `2026-06-09-container-runtime-module-boundary`
- Plan: `docs/plan/impl-plan/2026-06-09-container-runtime-module-boundary-implementation-plan.md`
- Diff source: 当前 worktree 相对 `HEAD` 的未提交改动（含未跟踪文件）
- Files reviewed:
  - `docs/plan/impl-plan/2026-06-09-container-runtime-module-boundary-implementation-plan.md`
  - `code/backend/internal/module/container_runtime/runtime/module.go`
  - `code/backend/internal/module/container_runtime/runtime/module_test.go`
  - `code/backend/internal/module/container_runtime/architecture_test.go`
  - `code/backend/internal/module/runtime/runtime/module.go`
  - `code/backend/internal/module/runtime/runtime/module_test.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/architecture_rules_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/module/architecture_baseline_test.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`

## Classification Check

- Agree with `非琐碎任务`。
- 理由：这次变更同时触达 module builder owner、app composition 装配、架构基线 allowlist、架构事实文档和迁移技术债表述，属于典型的 backend architecture gate review。

## Gate Verdict

- `blocked`

## Findings

### P1. 旧 `runtime/runtime` 并没有保持兼容转发，导出的 `Module` 字段面已经被破坏

- 计划在 `docs/plan/impl-plan/2026-06-09-container-runtime-module-boundary-implementation-plan.md:154-156` 明确要求“Keep old package API compiling by aliasing `Module`, `Deps`, and `BackgroundJob`, and forwarding `Build`”。
- 但当前 `code/backend/internal/module/runtime/runtime/module.go:3-10` 直接把 `Module` alias 到 `containerruntime.Module`，而新类型在 `code/backend/internal/module/container_runtime/runtime/module.go:20-33` 只导出 `ContainerFiles`，不再导出旧表面的 `ContestContainerFiles`。
- 这不是纯内部重命名，因为旧包的兼容层正是为“保留旧 import path 继续可编译”而存在；现在任何仍按旧字段名访问 `runtimemodule.Module.ContestContainerFiles` 的遗留调用方都会直接编译失败。连当前回归测试也被同步改成了新字段名 `module.ContainerFiles`（`code/backend/internal/module/runtime/runtime/module_test.go:18`），说明 compat surface 已经实际收缩。
- 影响：这会把“旧包只兼容转发”的承诺变成“旧包路径还在，但 exported shape 已变”，后续如果还有未迁完的旧调用点或外部 review 以为这里是稳定 compat layer，就会在再次接入或 cherry-pick 时踩到静态编译错误。
- 修正方向：旧 `runtime/runtime` 要么保留一个显式 wrapper 结构，把 `ContestContainerFiles` 等旧导出名继续桥接到新 builder；要么明确放弃 compat 承诺，同时同步收缩 plan / 架构文档 / TODO 的口径，并补一个 compile guard 证明“允许破坏的兼容面只有哪些”。

### P1. 新增的 `container_runtime` guardrail 对 Docker SDK 的封锁不成立，文档和计划存在过度声明

- 计划在 `docs/plan/impl-plan/2026-06-09-container-runtime-module-boundary-implementation-plan.md:158-160` 写的是阻止新包导入 “Gin, GORM, Redis, Docker SDK, or module API packages”。
- 但当前 `code/backend/internal/module/container_runtime/architecture_test.go:43-46` 只把 `github.com/docker/docker` 这个根路径作为 blocked import；而实际 helper `assertFileDoesNotImport` 在 `code/backend/internal/module/container_runtime/architecture_test.go:63-64` 用的是字符串全等比较。
- 这意味着最常见的 Docker SDK 具体入口，比如 `github.com/docker/docker/client`、`github.com/docker/docker/api/types`、`github.com/docker/docker/errdefs`，都不会被这条 guard 挡住，测试仍会通过。
- 影响：这次切片的核心价值之一就是把 `container_runtime` 的底层 builder 边界先用 guardrail 固化下来；如果 Docker concrete import 还能从最常见子包路径直接绕过，当前 guard 就不足以支撑文档里“已阻止 Docker SDK concrete type 回流”的说法。
- 修正方向：把 blocked import 检查改成前缀匹配或复用现有 `archtest.ImportPathMatches` 一类 helper，至少覆盖 `github.com/docker/docker` 及其子包；修完后再保留当前文档口径。

## Material Findings

- 旧 `runtime/runtime` 的兼容层没有保持旧 exported `Module` 字段面，和本次切片的 compat 承诺冲突。
- `container_runtime` 的新增边界测试没有真正封住常见 Docker SDK 子包，guardrail 质量不足以支撑“底层 builder 已被严格隔离”的说法。

## Non-blocking Suggestions

- 在 compat 层修复后，补一个专门面向旧 import path 的 compile-only regression test，锁定旧导出字段和 `Build(...)` 返回 shape，避免后续迁移过程中再次无意收缩 compat surface。

## Missing Validation

- 当前没有任何 compile guard 能证明“旧 `internal/module/runtime/runtime` 对遗留调用方仍然保持原有 exported shape”；现有测试只证明新字段名下的 builder 能工作。
- 当前也没有一条 reviewer-side 或 self-check 证明 `container_runtime` 的 boundary test 能拦住 Docker SDK 常见子包路径；现有通过结果只能说明 exact-match 版本的检查通过了。

## Open Questions Or Assumptions

- 我按 plan 和架构文档的当前表述，把“旧 `runtime/runtime` 仅兼容转发”理解为：旧 import path 在这次切片之后仍应保留既有静态编译表面。若实现方本意是允许字段级 breaking change，需要先改文档和 plan，再把这条兼容性降级为显式非目标。
- 我把 `runtime <-> container_runtime` 的临时双向依赖视为本切片可接受的过渡状态，因为文档和 TODO 已明确写出这不是终态 owner；本次 blocking 点不在这条过渡本身，而在兼容口径和 guardrail 质量没有跟上。

## Senior Implementation Assessment

- 先把物理 module builder 落到 `internal/module/container_runtime/runtime`，再把 `runtime/{application,contracts,ports,infrastructure}` 留作过渡基座，这个拆法本身是合理的最小切片。它避免了把 builder 切换、能力迁移、repo owner 收口混成一大步。
- 当前主要问题不在切片方向，而在两个“承诺层”没有收紧到位：
  - compat 层说自己只是转发，但 exported shape 已经变了；
  - guardrail 说自己挡住 Docker SDK concrete import，但实现只挡住了根路径字符串。
- 这两个问题都会让后续 reviewer 和实现者对边界状态形成过强假设，所以应该在进入 completion 之前补齐。

## Required Re-validation

- 修复 compat surface 后，至少重跑：
  - `cd code/backend && go test ./internal/module/runtime/runtime -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestComposition' -count=1`
- 修复 guardrail 后，至少重跑：
  - `cd code/backend && go test ./internal/module/container_runtime/... -count=1`
  - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestCrossModulePrivateImportExceptionsAreCurrent|TestModuleArchitectureBoundaries' -count=1`
- 如果这两处修复改动到完成门禁脚本涉及的 allowlist 或 review exceptions，`completion-full` 也需要重新执行一次，原先的通过结果不能直接沿用。

## Residual Risk

- 我这轮没有重跑用户已提供的全部 suite，只复核了与 blocker 最相关的最小集合，并按静态代码阅读判断 architecture/doc promises 是否一致。
- 即使修完这轮 blocker，`container_runtime -> runtime` 和 `runtime -> container_runtime` 的过渡双向依赖仍然是后续迁移成本来源；只是当前文档已经把它标成了过渡基座，而不是被误写成最终边界。

## Touched Known-debt Status

- 本次 touched surface 直接命中 `docs/todos/2026-05-17-project-tech-debt-from-migrations.md` 里关于 runtime/container_runtime 边界迁移的活动技术债。
- 当前状态：`blocked by compat-surface drift and insufficient guardrail coverage`
- 这次不能把上述问题降成 residual risk，因为它们正好发生在本轮宣称“已完成底层模块边界第一刀”的 touched surface 上。

## Implementation Follow-up Resolution

Resolution date: 2026-06-10

Implementation-side status: material findings fixed and impacted validation rerun.

- P1 compat surface: fixed by changing old `internal/module/runtime/runtime` from a direct type alias into an explicit wrapper that preserves the legacy `ContestContainerFiles` exported field while delegating construction to `container_runtime/runtime.Build`.
- P1 guardrail coverage: fixed by changing the `container_runtime` architecture guard to reject blocked imports by exact match or subpackage prefix, covering Docker SDK subpackages such as `github.com/docker/docker/client`.

Validation rerun after fixes:

```bash
cd code/backend && timeout 180s go test ./internal/module/runtime/runtime -count=1
cd code/backend && timeout 180s go test ./internal/module/container_runtime/... -count=1
cd code/backend && timeout 180s go test ./internal/app/composition -run 'TestBuildContainerRuntimeModule|TestComposition' -count=1
cd code/backend && timeout 180s go test ./internal/module -run 'TestModuleDependencyBaselineIsCurrent|TestCrossModulePrivateImportExceptionsAreCurrent|TestModuleArchitectureBoundaries' -count=1
cd code/backend && timeout 180s go test ./internal/app -run 'TestArchitectureRulesConcreteCrossModuleImportExceptionsAreCurrent|TestArchitectureRulesRejectConcreteCrossModuleImports|TestRouterCompositionStructure' -count=1
timeout 900s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full
```

Result: all commands passed.
