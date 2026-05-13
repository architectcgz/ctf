# Contest AWD Query Round Team Lookup Contract Phase 5 Slice 34 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `contest/application/queries/awd_support.go` 与 `contest/application/queries/awd_workspace_query.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，把 AWD query path 的 round/team lookup not-found 语义收口到 `contest/ports` 和 query-only adapter。

**Architecture:** `contest/infrastructure` 新增一条 query-only AWD adapter，负责把 raw `AWDRepository` 暴露出来的 round/team `gorm.ErrRecordNotFound` 映射成模块内 sentinel；`contest/application/queries` 只消费模块语义错误；runtime 只在 query `AWDService` / readiness query 这条路径注入 adapter，command `AWDService` 保持 raw `awdRepo`。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `contest/application/queries/awd_support.go -> gorm.io/gorm`
- 删除 `contest/application/queries/awd_workspace_query.go -> gorm.io/gorm`
- team lookup 缺失继续复用 `contest/ports.ErrContestUserTeamNotFound`
- round lookup not-found 收口到 `contest/ports` sentinel，并保持现有 query 对外行为：
  - 指定 round 查询映射成 `errcode.ErrNotFound`
  - 当前运行 round 缺失时继续返回正常空状态

## Non-goals

- 不修改 `code/backend/internal/module/architecture_allowlist_test.go`
- 不修改 `docs/architecture/backend/07-modular-monolith-refactor.md`
- 不修改 `docs/design/backend-module-boundary-target.md`
- 不修改 AWD command / jobs 文件
- 不扩展到 query path 之外的其他 `gorm.ErrRecordNotFound` 清理

## Inputs

- `code/backend/internal/module/contest/application/queries/awd_support.go`
- `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- `code/backend/internal/module/contest/application/queries/awd_service.go`
- `code/backend/internal/module/contest/application/queries/awd_service_test.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_team_relation_repository.go`
- `code/backend/internal/module/contest/infrastructure/team_query_adapter.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/ports/team.go`

## Ownership Boundary

- `contest/application/queries/awd_support.go`
  - 负责：contest/round 查询前置校验与 errcode 映射
  - 不负责：知道底层 round not-found 来自 `gorm`
- `contest/application/queries/awd_workspace_query.go`
  - 负责：workspace 空状态、当前轮次缺失和我的队伍缺失的 query 语义
  - 不负责：知道底层 round/team not-found 来自 `gorm`
- `contest/ports/awd.go`
  - 负责：定义 AWD round lookup 的模块内 sentinel
  - 不负责：决定 HTTP / API 错误码
- `contest/infrastructure/awd_query_repository.go`
  - 负责：把 raw `AWDRepository` 的 round/team concrete not-found 映射成模块 sentinel
  - 不负责：改变其他 query 的业务逻辑或扩展 query surface
- `contest/runtime/module.go`
  - 负责：只给 query `AWDService` / readiness query 注入 adapter
  - 不负责：改动 command `AWDService` 或 jobs 依赖

## Change Surface

- Add: `.harness/reuse-decisions/contest-awd-query-round-team-lookup-contract-phase5-slice34.md`
- Add: `docs/plan/impl-plan/2026-05-13-contest-awd-query-round-team-lookup-contract-phase5-slice34-implementation-plan.md`
- Add: `code/backend/internal/module/contest/infrastructure/awd_query_repository.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_query_repository_test.go`
- Add: `code/backend/internal/module/contest/application/queries/awd_query_contract_test.go`
- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_support.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

## Task 1: 先写失败测试

**Files:**
- Add: `code/backend/internal/module/contest/infrastructure/awd_query_repository_test.go`
- Add: `code/backend/internal/module/contest/application/queries/awd_query_contract_test.go`

- [ ] 为 query-only adapter 补 round/team not-found 映射测试
- [ ] 为 AWD query application 补 round/team not-found contract 测试
- [ ] 跑最小测试，确认红灯来自缺失的 sentinel / adapter / wiring

验证：

- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'AWDQueryRepository' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/application/queries -run 'AWDService.*(Round|Workspace).*NotFound' -count=1 -timeout 5m`

Review focus：

- adapter 测试是否只约束 concrete not-found 到模块 sentinel 的窄转换
- application query 测试是否在约束对外语义，而不是继续耦合 `gorm`

## Task 2: 实现 query-only adapter 与 query 侧 contract

**Files:**
- Add: `code/backend/internal/module/contest/infrastructure/awd_query_repository.go`
- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_support.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

- [ ] 在 `contest/ports/awd.go` 增加 AWD round lookup not-found sentinel
- [ ] 新增 query-only adapter，映射 `FindRoundByContestAndID`、`FindRunningRound`、`FindContestTeamByMember`
- [ ] 让 query application 只消费模块 sentinel，删除 `gorm` import
- [ ] 在 runtime 和 query helper 里只给 query `AWDService` 注入 adapter

验证：

- `cd code/backend && go test ./internal/module/contest/application/queries -run 'AWDService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'AWDQueryRepository' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`

Review focus：

- `awd_support.go` 和 `awd_workspace_query.go` 是否已经完全脱离 `gorm` concrete
- runtime 是否只在 query 路径注入 adapter，而 command/jobs 仍保留 raw `awdRepo`
- 当前运行 round 缺失和用户未加入队伍时，workspace 是否仍保持既有空状态语义

## Risks

- `AWDService` 既服务 readiness query，也服务 workspace/round 查询；如果 runtime 注入遗漏，测试 helper 可能通过但真实 wiring 仍会绕过 mapping
- `GetUserWorkspace` 同时容忍 current round 缺失和 my team 缺失；如果错误映射过宽，可能把非 not-found 异常误吞成空状态
- 受本 slice 范围限制，这次只处理 query path，不处理 AWD command/jobs 中仍存在的 raw `gorm` 分支

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -run 'AWDQueryRepository' -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/queries -run 'AWDService' -count=1 -timeout 300s`
3. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`

## Architecture-Fit Evaluation

- owner 明确：not-found concrete 由 `contest/infrastructure + contest/ports` 收口，application query 只决定 errcode / 空返回
- reuse point 明确：复用 phase5 已有的 “局部 sentinel + 窄 adapter” 模式，不改 raw repository 全局语义
- 结构收敛明确：本 slice 只收 AWD query round/team lookup，不扩展到 command/jobs 或 shared allowlist/docs
