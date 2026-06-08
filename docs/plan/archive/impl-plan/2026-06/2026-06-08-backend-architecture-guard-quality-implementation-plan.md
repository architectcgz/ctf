<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 2026-06-08-backend-architecture-guard-quality Implementation Plan

**Goal:** 优化后端架构守卫的可维护性，并补上 `internal/shared` 的 owner 边界守卫。

**Architecture:** 新增 `internal/testutil/archtest` 作为后端架构测试 helper owner，承接 Go 文件收集、import 解析、文件读取和 import 禁止断言；先改造全局 module/app 架构守卫复用该 helper。新增 `internal/shared/architecture_test.go`，明确 shared 只能承接跨 bounded context 的轻量基础能力，不能反向依赖 `internal/module/*`、app/composition 或 Redis/GORM/Docker/Gin 等具体基础设施。

**Tech Stack:** Go AST parser, Go test, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-08-backend-architecture-guard-quality`
- Started At: `2026-06-08T12:36:29Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-backend-architecture-guard-quality`
- Branch: `task/2026-06-08-backend-architecture-guard-quality`

## Objective And Non-Goals

- Objective:
  - 抽出后端架构测试公共 helper，减少全局架构守卫里重复的文件扫描和 import 解析代码。
  - 补 `internal/shared` 边界守卫，防止 shared 演化成业务 / 基础设施混杂目录。
  - 保持现有 `scripts/check-backend-architecture.sh --full` 接入方式不变。
- Non-Goals:
  - 不一次性重写所有模块级 `architecture_test.go`。
  - 不改变现有业务代码、模块依赖 baseline 或 workflow stage 注册。
  - 不把架构守卫从 Go test 迁移到 Python。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
- Related architecture/contracts:
  - `scripts/check-backend-architecture.sh`
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/app/architecture_rules_test.go`
  - `code/backend/internal/app/backend_context_architecture_test.go`
- Related prior work:
  - 用户确认 `lockkeepalive` 当前可放在 `internal/shared`，但 shared 需要守住轻量、无业务语义、无基础设施实现依赖的边界。
  - `270c6b2da fix(backend): 收口多实例后台 owner 治理` 已完成旧 keepalive、runtime cleaner、assessment cleaner 的多副本 owner 复核，并拆分 `tools/multi-instance-nginx-proxy-smoke.sh`；当前任务分支需要带入该依赖提交，否则 `workflow-governance` 会被旧脚本长度 guard 阻断。

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 触达后端架构守卫和测试基础设施，会影响 workflow gate 质量；需要计划、验证和独立 review。

## Files

- Create:
  - `code/backend/internal/testutil/archtest/archtest.go`
  - `code/backend/internal/shared/architecture_test.go`
- Modify:
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
- Review:
  - `code/backend/internal/app/backend_context_architecture_test.go`
  - `code/backend/internal/module/*/architecture_test.go`
  - `code/backend/internal/shared/*`
- Test:
  - `go test ./internal/testutil/archtest`
  - `go test ./internal/shared -run TestSharedPackagesStayLightweightAndModuleAgnostic -count=1`
  - `go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestModuleDependencyBaselineIsCurrent|TestCrossModulePrivateImportExceptionsAreCurrent|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1`
  - `go test ./internal/app -run TestArchitectureRulesRejectConcreteCrossModuleImports -count=1`
  - `bash scripts/check-backend-architecture.sh --full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`

## 复用与 Owner 决策

- Existing patterns searched:
  - 现有守卫分布在 `internal/module/architecture_test.go`、各模块 `architecture_test.go`、`internal/app/*architecture*.go`、`tests/architecture`。
  - `internal/testutil` 已经承接测试辅助能力，适合新增架构测试 helper。
- Reuse / extend / split / create-new decision:
  - 新增 `internal/testutil/archtest`，不放进生产 `internal/shared`，避免把测试扫描能力误归到业务共享能力。
  - 本轮只迁移全局 module/app 守卫和同包 mapper policy 到 helper；各模块私有守卫保留，避免一次性大 diff。
- Owner boundary:
  - `archtest` 只做测试用 Go 文件和 import 读取/断言。
  - `internal/shared/architecture_test.go` 只定义 shared 的边界规则，不拥有其他模块的架构规则。
- Why this is the narrowest safe surface:
  - 能立即减少核心守卫重复，并补上当前明确缺口；不扰动业务代码和所有模块私有守卫。

## Implementation Notes

- `scripts/check-backend-architecture.sh` 的 module 阶段从只跑 `TestModuleArchitectureBoundaries` 调整为跑完整 `./internal/module` 架构包，避免 baseline / exception currentness、root context、`time.Now`、transaction 和 mapper policy 守卫只靠人工执行。
- 扩大脚本覆盖后暴露两个既有隐藏失败：
  - `internal/bootstrap/awd_defense_ssh_gateway.go` 是真实进程入口，已与 app context guard 对齐加入 module root context exception。
  - `awd_attack_flag_rotation_support.go` 不再直接调用 `time.Now()`，改用提交链路 `submitAttackContext` 已记录的 UTC 时间；`runtime/infrastructure/node_repository.go` 作为 repository 时间戳写入点加入 reviewed `time.Now` exception。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 这是架构守卫质量优化，需要先明确优化边界，避免把“优化”扩大成重写全部守卫。
- grill-with-docs findings:
  - 后端架构文档明确 `internal/shared` 是“映射与轻量共享工具”，不是业务 owner 或基础设施实现层。
  - `internal/module/*` 是主要 bounded-context surface，跨模块依赖已有 baseline / reviewed exception 守卫。
  - 当前缺少 `internal/shared` 自身不能依赖业务模块 / 基础设施实现的机械守卫。
- Plan adjustments after challenge:
  - 第一阶段只补 `archtest` helper 和 `internal/shared` 守卫；后续如果需要，再逐步迁移各模块私有 guard。

## Validation

- Commands:
  - `go test ./internal/testutil/archtest`
  - `go test ./internal/shared -count=1`
  - `go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestModuleDependencyBaselineIsCurrent|TestCrossModulePrivateImportExceptionsAreCurrent' -count=1`
  - `go test ./internal/app -run TestArchitectureRulesRejectConcreteCrossModuleImports -count=1`
  - `bash scripts/check-backend-architecture.sh --full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`
- Results:
  - `timeout 120s go test ./internal/testutil/archtest ./internal/shared ./internal/module ./internal/app -run 'TestSharedPackagesStayLightweightAndModuleAgnostic|TestModuleArchitectureBoundaries|TestModuleDependencyBaselineIsCurrent|TestDomainInternalImportExceptionsAreCurrent|TestApplicationConcreteDependencyExceptionsAreCurrent|TestCrossModulePrivateImportExceptionsAreCurrent|TestModuleRuntimeCodeDoesNotCreateRootContext|TestBackendBusinessCodeDoesNotCreateRootContext|TestTimeNowUsageExceptionsAreCurrent|TestTransactionBoundaryExceptionsAreCurrent|TestRuntimeModulesStaySmallAndWiringOnly|TestMapperWrappersFollowGlobalDelegationPolicy|TestArchitectureRulesRejectConcreteCrossModuleImports' -count=1`：通过。
  - `timeout 180s bash scripts/check-backend-architecture.sh --full`：通过。
  - `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`：通过。
  - `timeout 600s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`：通过。
  - `timeout 300s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`：通过。
- Manual checks:
  - `archtest` 不进入 `internal/shared`，避免污染业务 shared。
  - `internal/shared` guard 不误伤当前 `flagcrypto`、`lockkeepalive`、`taxonomy`。
- Review focus:
  - helper 抽取是否降低重复而不模糊失败信息。
  - `internal/shared` 规则是否过宽或过窄。
  - workflow 接入是否仍通过现有 backend architecture entrypoint。
