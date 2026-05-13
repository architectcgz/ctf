# Challenge Core Command Not-Found Contract Phase 5 Slice 35 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `challenge/application/commands/challenge_service.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持题目创建、更新、删除、发布自检和自测的既有业务语义不变。

**Architecture:** 新增一层窄 `ChallengeCommandRepository` adapter，把 raw challenge repository 的 challenge / publish-check lookup not-found 收口成 `challenge/ports` sentinel；core challenge command service 只消费模块 sentinel；runtime 只给 core command path 注入 adapted command/image/topology repo，package revision 继续保留 raw repo；仍需 `gorm.DB` 的构造与 struct 留在 `challenge_import_service.go`。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `challenge/application/commands/challenge_service.go -> gorm.io/gorm`
- 保持 `CreateChallenge` 的 image not-found 仍映射成 `errcode.ErrNotFound`
- 保持 `UpdateChallenge` / `DeleteChallenge` / `PublishChallenge` / `SelfCheckChallenge` 的 challenge not-found 仍映射成既有 errcode
- 保持 `RequestPublishCheck` 在“没有 active job”时继续创建新 job
- 保持 `GetLatestPublishCheck` 的 stale / missing job 分支语义不变

## Non-goals

- 不修改 `challenge/application/commands/challenge_package_revision_service.go`
- 不修改 `challenge/application/commands/image_build_service.go`
- 不修改 `challenge/application/commands/awd_challenge_import_service.go`
- 不改变 raw `challengeinfra.Repository` 的全局 not-found 语义
- 不处理 contest / practice / assessment / ops 模块

## Inputs

- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_error_contract_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_self_check_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Ownership Boundary

- `challenge/application/commands/challenge_service.go`
  - 负责：core challenge command 编排、既有 errcode 映射、自测与发布自检流程
  - 不负责：知道 challenge / publish-check not-found 是否来自 GORM
- `challenge/infrastructure/challenge_command_repository.go`
  - 负责：把 raw challenge repository 的 core command lookup not-found 映射成模块 sentinel
  - 不负责：改变 create/update/delete 与 publish-check 写入语义
- `challenge/application/commands/challenge_import_service.go`
  - 负责：保留 import / export 事务面需要的 `gorm.DB` 构造入口
  - 不负责：让 core challenge command surface 继续直接 import GORM
- `challenge/runtime/module.go`
  - 负责：只给 core command service 注入 adapted command/image/topology repo
  - 不负责：把 adapted 语义扩散到 import / package revision 等其他 use case

## Change Surface

- Add: `.harness/reuse-decisions/challenge-core-command-not-found-contract-phase5-slice35.md`
- Add: `docs/plan/impl-plan/2026-05-13-challenge-core-command-not-found-contract-phase5-slice35-implementation-plan.md`
- Add: `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
- Add: `code/backend/internal/module/challenge/infrastructure/challenge_command_repository_test.go`
- Add: `code/backend/internal/module/challenge/application/commands/challenge_error_contract_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_self_check_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`
- Modify: `code/backend/internal/app/challenge_import_integration_test.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/design/backend-module-boundary-target.md`

## Task 1: 合约测试与 adapter 测试先变红

**Files:**
- `code/backend/internal/module/challenge/application/commands/challenge_error_contract_test.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_command_repository_test.go`

- [x] 为 core challenge command 的 sentinel 分支补齐 application contract 测试
- [x] 为 raw GORM not-found -> challenge command sentinel 补齐 adapter 测试
- [x] 确认红灯来自 `challenge_service.go` 仍直接依赖 `gorm.ErrRecordNotFound`

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'ChallengeService.*(Module|Sentinel)|CreateChallengeTreatsModuleImageNotFound|RequestPublishCheckTreatsMissingActiveJobSentinel|GetLatestPublishCheckTreatsMissingJobSentinel|UpdateChallengeTreatsTopologySentinel' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'ChallengeCommandRepository' -count=1 -timeout 300s`

## Task 2: 实现 adapter、wiring 与构造挪位

**Files:**
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_self_check_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/app/challenge_import_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

- [x] 新增 `ChallengeCommandRepository` adapter，把 challenge / publish-check lookup not-found 映射成模块 sentinel
- [x] 让 `challenge_service.go` 只消费模块 sentinel，不再 import GORM
- [x] 把 `ChallengeService` struct / `NewChallengeService` / `SelfCheckConfig` 挪到仍允许依赖 GORM 的 `challenge_import_service.go`
- [x] 修正 stub 与 DB-based 构造点，统一注入 adapted command/image/topology repo
- [x] 保持 package revision repo 继续走 raw repo，不扩大这刀范围

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'ChallengeService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'Challenge(Command|Image(Command|Query)|TopologyService)Repository' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/app -run 'TestChallengeImport|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1 -timeout 300s`

## Task 3: shared cleanup 与事实源同步

**Files:**
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/design/backend-module-boundary-target.md`
- `.harness/reuse-decisions/challenge-core-command-not-found-contract-phase5-slice35.md`
- `docs/plan/impl-plan/2026-05-13-challenge-core-command-not-found-contract-phase5-slice35-implementation-plan.md`

- [x] 删除 `challenge/application/commands/challenge_service.go -> gorm.io/gorm` allowlist
- [x] 把 core challenge command not-found contract 的 owner、adapter 注入面和 remaining raw transaction surface 写回事实源
- [x] 补齐本 slice reuse decision 与 implementation plan

## Risks

- `RequestPublishCheck` 的 missing active job 语义是“允许创建”，不能被误映射成对外 not-found
- `GetLatestPublishCheck` 仍要区分 challenge 更新时间与历史 job 的 stale 语义，不能因为 adapter 化把 stale job 误当最新有效结果
- `ChallengeService` 构造挪位后，import / export 事务面仍要保留 `gorm.DB` 注入，不要被这刀误删
- DB-based 测试如果继续走 raw repo，会把旧的 `gorm.ErrRecordNotFound` 重新带回 command service

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'ChallengeService' -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'Challenge(Command|Image(Command|Query)|TopologyService)Repository' -count=1 -timeout 300s`
3. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s`
4. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/app -run 'TestChallengeImport|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1 -timeout 300s`
