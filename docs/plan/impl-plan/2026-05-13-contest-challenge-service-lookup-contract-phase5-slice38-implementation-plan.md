# Contest Challenge Service Lookup Contract Phase 5 Slice 38 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `contest/application/commands/challenge_add_commands.go` 与 `contest/application/commands/contest_awd_service_service.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，并保持 command 既有业务语义不变。

**Architecture:** 在 `contest/infrastructure` 增加 challenge catalog lookup adapter 与 AWD challenge lookup adapter，把 cross-module contract 暴露出来的 not-found 收口到 `contest/ports` sentinel；`ContestAWDServiceService.repo` 直接复用现有 `AWDCommandRepository` 的 service not-found 映射；runtime 只给这两个 command service 注入 adapter。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, contract tests

---

## Objective

- 删除 `contest/application/commands/challenge_add_commands.go -> gorm.io/gorm`
- 删除 `contest/application/commands/contest_awd_service_service.go -> gorm.io/gorm`

## Non-goals

- 不修改 `code/backend/internal/module/architecture_allowlist_test.go`
- 不修改 `docs/architecture/backend/07-modular-monolith-refactor.md`
- 不修改 `docs/design/backend-module-boundary-target.md`
- 不修改 `contest/application/jobs/*`
- 不修改 `contest/application/commands/awd_*` 已收口的 slice37 面
- 不触碰 challenge image build / package revision / import 相关 challenge surface

## Inputs

- `.harness/reuse-decisions/contest-challenge-service-lookup-contract-phase5-slice38.md`
- `code/backend/internal/module/contest/application/commands/challenge_add_commands.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/commands/challenge_service.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/challenge.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/infrastructure/awd_command_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- `code/backend/internal/module/challenge/contracts/contracts.go`

## Ownership Boundary

- `contest/application/commands/challenge_add_commands.go`
  - 负责：把 contest challenge add 的缺失题目语义映射成 `errcode.ErrChallengeNotFound`
  - 不负责：知道缺失题目来自 GORM 或上游 challenge module raw repo
- `contest/application/commands/contest_awd_service_service.go`
  - 负责：把缺失 AWD challenge / 缺失 contest AWD service / 缺失 challenge entity 语义映射成既有 errcode
  - 不负责：知道上游 challenge / awd challenge / awd service lookup 的 concrete not-found 细节
- `contest/ports/challenge.go`
  - 负责：定义这刀需要的 challenge entity sentinel
- `contest/infrastructure/contest_challenge_lookup_adapter.go`
  - 负责：把 `challengecontracts.ContestChallengeContract.FindByID` 的 concrete not-found 收口
- `contest/infrastructure/contest_awd_challenge_lookup_adapter.go`
  - 负责：把 `challengeports.AWDChallengeQueryRepository.FindAWDChallengeByID` 的 concrete / upstream sentinel 收口
- `contest/runtime/module.go`
  - 负责：只给 `ChallengeService` command 与 `ContestAWDServiceService` 注入 adapter

## Change Surface

- Add: `.harness/reuse-decisions/contest-challenge-service-lookup-contract-phase5-slice38.md`
- Add: `docs/plan/impl-plan/2026-05-13-contest-challenge-service-lookup-contract-phase5-slice38-implementation-plan.md`
- Add: `code/backend/internal/module/contest/infrastructure/contest_challenge_lookup_adapter.go`
- Add: `code/backend/internal/module/contest/infrastructure/contest_challenge_lookup_adapter_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter.go`
- Add: `code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter_test.go`
- Add: `code/backend/internal/module/contest/application/commands/contest_challenge_error_contract_test.go`
- Modify: `code/backend/internal/module/contest/ports/challenge.go`
- Modify: `code/backend/internal/module/contest/application/commands/challenge_add_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/challenge_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

## Task 1: 先写失败测试

**Files:**

- Add: `code/backend/internal/module/contest/application/commands/contest_challenge_error_contract_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/contest_challenge_lookup_adapter_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter_test.go`

- [ ] 为 `ChallengeService` 缺失 challenge entity -> `ErrChallengeNotFound` 补 contract test
- [ ] 为 `ContestAWDServiceService` 缺失 AWD challenge / contest awd service -> `ErrNotFound` 补 contract test
- [ ] 为两个 infrastructure adapter 补 concrete / upstream sentinel -> contest sentinel 映射测试

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/commands -run 'ChallengeService|ContestAWDServiceService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -run 'Contest(ChallengeLookupAdapter|AWDChallengeLookupAdapter)' -count=1 -timeout 300s`

Review focus：

- contract 测试是否在约束 contest sentinel，而不是继续借 `gorm.ErrRecordNotFound` 过关
- infrastructure 测试是否覆盖 upstream `challengeports.ErrAWDChallengeNotFound` 映射，而不是只测 raw gorm

## Task 2: 实现 adapter、ports sentinel 与 runtime wiring

**Files:**

- Modify: `code/backend/internal/module/contest/ports/challenge.go`
- Add: `code/backend/internal/module/contest/infrastructure/contest_challenge_lookup_adapter.go`
- Add: `code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter.go`
- Modify: `code/backend/internal/module/contest/application/commands/challenge_add_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/challenge_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

- [ ] 在 `contest/ports/challenge.go` 增加 challenge entity not-found sentinel
- [ ] 新增 `ContestChallengeLookupAdapter`，收口 `FindByID`
- [ ] 新增 `ContestAWDChallengeLookupAdapter`，收口 `FindAWDChallengeByID`
- [ ] `contest_awd_service_service.go` 复用现有 `AWDCommandRepository` 的 service not-found 映射，不新增重复 service adapter
- [ ] 让两个 command service 只消费 contest sentinel，删除 `gorm` import
- [ ] runtime 只改 command wiring，不碰 query / jobs / slice37 已落地 AWDService wiring

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/commands -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`

Review focus：

- `ChallengeService` command 与 `ContestAWDServiceService` 是否已完全脱离 `gorm` concrete
- runtime 是否只动这两个 command service 的 wiring
- `ContestAWDServiceService` 是否正确兼容 upstream `challengeports.ErrAWDChallengeNotFound`

## Risks

- `ChallengeCatalog` 与 `AWDChallengeQueryRepo` 都来自 challenge module composition；如果 adapter 只映射 raw `gorm.ErrRecordNotFound`，真实 runtime 仍可能漏掉 challenge module 已经转换过的 sentinel
- `syncContestChallengeRelation` 目前是 dormant 分支，容易被忽略，但它和 `challenge_add_commands.go` 共享同一 challenge entity lookup 语义
- 如果把 cross-module challenge lookup 继续塞进 `AWDCommandRepository`，会让现有 AWD command adapter 边界变宽

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/commands -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -count=1 -timeout 300s`
3. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`

## Architecture-Fit Evaluation

- owner 明确：challenge entity lookup 与 AWD challenge lookup concrete 由 `contest/infrastructure` 收口，command application 只决定 errcode
- reuse point 明确：service not-found 继续复用 `AWDCommandRepository`，不重复造 service adapter；cross-module challenge lookup 走单独窄 adapter
- 结构收敛明确：只处理 `challenge_add_commands.go` 与 `contest_awd_service_service.go` 这两条剩余 contest command gorm allowlist，不扩到 jobs、query 或 challenge module 其它 surfaces
