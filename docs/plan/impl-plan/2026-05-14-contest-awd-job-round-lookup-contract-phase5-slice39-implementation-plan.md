# Contest AWD Job Round Lookup Contract Phase 5 Slice 39 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `contest/application/jobs/awd_check_cache_support.go`、`contest/application/jobs/awd_round_flag_lookup_support.go` 与 `contest/application/jobs/awd_round_runtime.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，并同步收口 `AWDRoundUpdater` 的 runtime 注入路径。

**Architecture:** 在 `contest/infrastructure` 增加 job-only round lookup adapter，把 `AWDRoundUpdater` 使用的 `FindRunningRound` / `FindRoundByNumber` concrete not-found 收口到 `contestports.ErrContestAWDRoundNotFound`；`EnsureActiveRoundMaterialized` 直接返回该 sentinel；runtime 只给 job 路径注入 adapted repo。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, job contract tests

---

## Objective

- 删除 `contest/application/jobs/awd_check_cache_support.go -> gorm.io/gorm`
- 删除 `contest/application/jobs/awd_round_flag_lookup_support.go -> gorm.io/gorm`
- 删除 `contest/application/jobs/awd_round_runtime.go -> gorm.io/gorm`

## Non-goals

- 不修改 `contest/application/jobs/*net/http*` 相关文件
- 不修改 `challenge/*` 剩余 allowlist
- 不修改 `contest/application/commands/*` 已完成 slice
- 不修改长期设计文档

## Inputs

- `.harness/reuse-decisions/contest-awd-job-round-lookup-contract-phase5-slice39.md`
- `code/backend/internal/module/contest/application/jobs/awd_check_cache_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_runtime.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- `code/backend/internal/module/contest/infrastructure/awd_command_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_query_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/awd.go`

## Ownership Boundary

- `contest/application/jobs/awd_check_cache_support.go`
  - 负责：把 “当前运行 round 不存在” 解释成 fallback 到 state store
  - 不负责：知道 not-found 来自 GORM
- `contest/application/jobs/awd_round_flag_lookup_support.go`
  - 负责：把 “上一轮不存在” 解释成只接受当前轮 flag
  - 不负责：知道 not-found 来自 GORM
- `contest/application/jobs/awd_round_runtime.go`
  - 负责：把 “当前时间没有活跃 round” 表达成 module sentinel
  - 不负责：向上游泄漏 GORM sentinel
- `contest/infrastructure/awd_job_repository.go`
  - 负责：把 `AWDRoundUpdater` 需要的 round lookup concrete not-found 收口到 `ErrContestAWDRoundNotFound`
- `contest/runtime/module.go`
  - 负责：只给 `AWDRoundUpdater` 注入 adapted repo

## Change Surface

- Add: `.harness/reuse-decisions/contest-awd-job-round-lookup-contract-phase5-slice39.md`
- Add: `docs/plan/impl-plan/2026-05-14-contest-awd-job-round-lookup-contract-phase5-slice39-implementation-plan.md`
- Add: `code/backend/internal/module/contest/infrastructure/awd_job_repository.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_job_repository_test.go`
- Add: `code/backend/internal/module/contest/application/jobs/awd_round_contract_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_check_cache_support.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_runtime.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_testsupport_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`

## Task 1: 先写 contract tests

**Files:**

- Add: `code/backend/internal/module/contest/application/jobs/awd_round_contract_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_job_repository_test.go`

- [ ] 为 `shouldSyncLiveServiceStatusCache` 缺失 running round -> fallback state store 补 contract test
- [ ] 为 `resolveAcceptedRoundFlags` 缺失 previous round -> 只保留当前轮 flag 补 contract test
- [ ] 为 `EnsureActiveRoundMaterialized` 无 active round -> `ErrContestAWDRoundNotFound` 补 contract test
- [ ] 为 job adapter 补 `gorm.ErrRecordNotFound -> ErrContestAWDRoundNotFound` 映射测试

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -run 'AWDRound.*(Cache|Flag|Materialized)' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -run 'AWD(JobRepository)' -count=1 -timeout 300s`

Review focus：

- job contract tests 是否约束 contest sentinel 语义，而不是继续借 `gorm.ErrRecordNotFound` 过关
- adapter tests 是否只做 not-found translation，不改变其他行为

## Task 2: 实现 adapter、jobs sentinel 消费与 runtime wiring

**Files:**

- Add: `code/backend/internal/module/contest/infrastructure/awd_job_repository.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_check_cache_support.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_runtime.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_testsupport_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`

- [ ] 新增 `AWDJobRepository`，只收口 `AWDRoundUpdater` 使用到的 round lookup not-found
- [ ] `shouldSyncLiveServiceStatusCache` 改为消费 `ErrContestAWDRoundNotFound`
- [ ] `resolveAcceptedRoundFlags` 改为消费 `ErrContestAWDRoundNotFound`
- [ ] `EnsureActiveRoundMaterialized` 直接返回 `ErrContestAWDRoundNotFound`
- [ ] runtime / tests helper 改成给 `AWDRoundUpdater` 注入 adapted repo，而不是 raw repo
- [ ] allowlist 删掉这 3 条 gorm 依赖

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s`

Review focus：

- `AWDRoundUpdater` 是否已完全脱离 `gorm` concrete round not-found
- runtime 是否只动 `AWDRoundUpdater` 的 wiring
- touched surface 上的 raw repo 注入 debt 是否已经同步收口

## Risks

- 如果复用过宽的 query adapter，而不是单独的 job adapter，会把 job debt 继续藏在 query 路径里
- `EnsureActiveRoundMaterialized` 的 not-found 语义同时被 command 侧 round manager adapter 消费；返回值调整后需要确认现有 adapter / tests 仍兼容
- tests helper 如果继续注 raw repo，会让应用层 contract 测试绕过真实 composition path

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -count=1 -timeout 300s`
3. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`
4. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s`

## Architecture-Fit Evaluation

- owner 明确：round lookup concrete 由 `contest/infrastructure/awd_job_repository.go` 收口，jobs application 只决定 fallback 语义
- reuse point 明确：`ErrContestAWDRoundNotFound` 继续复用既有 sentinel，不新增重复 round not-found 契约
- 结构收敛明确：这刀同时收口 jobs 文件里的 concrete import 和 `buildAWDHandler` 对 raw `AWDRepository` 的直接注入
