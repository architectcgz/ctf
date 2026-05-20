# Contest AWD Preview Runtime Lookup Not-Found Phase 5 Slice 32 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `contest/application/commands/awd_preview_runtime_support.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时把 AWD preview runtime lookup 的 not-found 语义收口到 `contest/ports` sentinel，并保留显式 URL 下现有的部分降级放行行为。

**Architecture:** `contest` 新增一条窄 preview runtime lookup adapter，把 `AWDChallengeQueryRepository` 与 `ImageStore` 暴露出来的 `gorm.ErrRecordNotFound` 映射成 `contest/ports` sentinel；`awd_preview_runtime_support.go` 只消费模块语义错误，再按现有规则映射到 `errcode` 或显式 URL 降级路径；runtime builder 负责注入 adapter。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `contest/application/commands/awd_preview_runtime_support.go -> gorm.io/gorm`
- 保持 preview challenge lookup not-found 时，显式 URL 仍可按现有逻辑降级放行
- 保持 runtime image lookup not-found 时，显式 URL 仍按现有逻辑阻断并返回既有错误语义

## Non-goals

- 不修改 `code/backend/internal/module/architecture_allowlist_test.go`
- 不修改 `docs/architecture/backend/07-modular-monolith-refactor.md`
- 不修改 `docs/design/backend-module-boundary-target.md`
- 不修改 shared allowlist / shared facts docs
- 不处理 `contest/application/commands/awd_service_run_commands.go`、其他 AWD command/query/jobs 或 challenge 模块代码

## Inputs

- `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/challenge/contracts/contracts.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Ownership Boundary

- `contest/application/commands/awd_preview_runtime_support.go`
  - 负责：AWD preview runtime lookup 的 errcode 映射、显式 URL 降级和自动拉起前置校验
  - 不负责：知道底层 not-found 是否来自 `gorm`
- `contest/ports/awd.go`
  - 负责：定义 AWD preview runtime lookup 的模块内 sentinel
  - 不负责：决定对外 HTTP / API 错误码
- `contest/infrastructure/awd_preview_runtime_lookup_repository.go`
  - 负责：把 challenge query/image store 暴露出来的 not-found 映射成 contest sentinel
  - 不负责：改变 preview checker 业务逻辑、runtime 启动或 token 生成语义
- `contest/runtime/module.go`
  - 负责：给 `AWDService` 注入 preview runtime lookup adapter
  - 不负责：把 raw GORM concrete 再带回 application surface

## Change Surface

- Add: `.harness/reuse-decisions/contest-awd-preview-runtime-lookup-not-found-phase5-slice32.md`
- Add: `docs/plan/impl-plan/2026-05-13-contest-awd-preview-runtime-lookup-not-found-phase5-slice32-implementation-plan.md`
- Add: `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository_test.go`
- Add: `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support_contract_test.go`
- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

## Task 1: 先写失败测试

**Files:**
- Add: `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support_contract_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository_test.go`

- [ ] 为 preview runtime support 补 sentinel 分支测试，先约束显式 URL 下 challenge/image not-found 的差异化行为
- [ ] 为 preview runtime lookup adapter 补 `AWDChallengeQueryRepository` / `ImageStore` not-found 映射测试
- [ ] 跑最小测试，确认红灯来自目标 contract 缺失

验证：

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'AWDService.*Preview.*(ExplicitURL|ImageRef)' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'AWDPreviewRuntimeLookup' -count=1 -timeout 5m`

Review focus：

- application 测试是否真正在约束 contest sentinel，而不是继续借 `gorm` sentinel 过关
- adapter 测试是否只覆盖 not-found 映射，不夹带 AWD preview 业务逻辑

## Task 2: 实现 sentinel 映射与 runtime wiring

**Files:**
- Add: `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

- [ ] 在 `contest/ports/awd.go` 增加 preview runtime lookup not-found sentinel
- [ ] 新增 contest infrastructure adapter，承接 challenge/image lookup 的 `gorm.ErrRecordNotFound`
- [ ] 让 `awd_preview_runtime_support.go` 只消费 contest sentinel，并保留显式 URL 的既有降级边界
- [ ] 在 runtime wiring 中注入 adapter，避免运行时继续绕过 mapping

验证：

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'AWDService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'AWD|Preview' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`

Review focus：

- `awd_preview_runtime_support.go` 是否已经完全去掉 `gorm` concrete
- challenge/image not-found 是否仍保持“部分降级、部分阻断”的既有语义

## Risks

- `prepareCheckerPreviewAccessURL` 同时依赖 challenge lookup 与 image lookup；如果把 not-found 统一映射成同一条上层行为，会破坏显式 URL 的放行边界
- `AWDService` 的测试 helper 若继续直接注入 raw repository，测试可能无法覆盖 runtime wiring 的真实行为
- 因用户明确禁止触碰 allowlist 与 shared docs，本 slice 只交付代码侧 contract 收口与局部验证证据

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/commands -run 'AWDService' -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -run 'AWD|Preview' -count=1 -timeout 300s`
3. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`

## Architecture-Fit Evaluation

- owner 明确：application 只保留 errcode / 降级映射，not-found concrete 由 `contest/ports + contest/infrastructure` 收口
- reuse point 明确：复用现有 `challenge` raw repository/store，新增一层 contest-side adapter，而不是改 challenge 模块
- 结构收敛明确：本 slice 只处理 AWD preview runtime lookup，不扩到其他 AWD command/query/jobs 或 shared allowlist/docs
