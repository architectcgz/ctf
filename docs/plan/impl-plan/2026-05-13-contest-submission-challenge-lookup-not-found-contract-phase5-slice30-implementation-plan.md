# Contest Submission Challenge Lookup Not-Found Contract Phase 5 Slice 30 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `contest/application/commands/submission_submit_validation.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时把 contest submission challenge lookup 的 not-found 语义收口到 `contest/ports` sentinel，不触碰 shared allowlist / shared facts docs。

**Architecture:** 继续复用现有 `SubmissionRegistrationRepository` 作为 submission wiring adapter，在 `contest/ports/submission.go` 增加 challenge lookup sentinel，把 raw `SubmissionRepository` 的 `FindContestChallenge` / `FindChallengeByID` not-found 映射成模块语义；`SubmissionService` 只负责 sentinel -> `errcode` 映射。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `contest/application/commands/submission_submit_validation.go -> gorm.io/gorm`
- 保持 contest challenge 缺失时继续返回 `errcode.ErrChallengeNotInContest`
- 保持 challenge 实体缺失时继续返回 `errcode.ErrChallengeNotFound`

## Non-goals

- 不修改 `code/backend/internal/module/architecture_allowlist_test.go`
- 不修改 `docs/architecture/backend/07-modular-monolith-refactor.md`
- 不修改 `docs/design/backend-module-boundary-target.md`
- 不修改 shared allowlist / shared facts docs
- 不处理 `submission_scoring.go`、AWD support / jobs / query 的其他 not-found concrete
- 不改 challenge 模块代码

## Inputs

- `code/backend/internal/module/contest/application/commands/submission_submit_validation.go`
- `code/backend/internal/module/contest/application/commands/submission_service.go`
- `code/backend/internal/module/contest/application/commands/participation_error_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/submission_lookup_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/submission.go`
- `docs/plan/impl-plan/2026-05-13-contest-participation-registration-not-found-contract-phase5-slice25-implementation-plan.md`

## Ownership Boundary

- `contest/application/commands/submission_submit_validation.go`
  - 负责：contest submission 前置校验和公开 `errcode` 映射
  - 不负责：知道 challenge lookup not-found 是否来自 `gorm`
- `contest/ports/submission.go`
  - 负责：定义 submission challenge lookup 的模块内 sentinel
  - 不负责：决定 HTTP / API 对外错误码
- `contest/infrastructure/submission_registration_repository.go`
  - 负责：把 raw submission repository 的 challenge lookup not-found 映射成 submission sentinel
  - 不负责：改变 submission scoring、写入或事务语义
- `contest/runtime/module.go`
  - 负责：继续把 adapter 注入 `SubmissionService`
  - 不负责：把 raw GORM concrete 带回 application surface

## Change Surface

- Add: `.harness/reuse-decisions/contest-submission-challenge-lookup-not-found-contract-phase5-slice30.md`
- Add: `docs/plan/impl-plan/2026-05-13-contest-submission-challenge-lookup-not-found-contract-phase5-slice30-implementation-plan.md`
- Modify: `code/backend/internal/module/contest/ports/submission.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_submit_validation.go`
- Modify: `code/backend/internal/module/contest/application/commands/participation_error_contract_test.go`
- Modify: `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- Modify: `code/backend/internal/module/contest/infrastructure/submission_registration_repository_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/contest/application/commands/participation_error_contract_test.go`
- Modify: `code/backend/internal/module/contest/infrastructure/submission_registration_repository_test.go`

- [ ] 为 `validateContestSubmission` 补 contest challenge / challenge entity 的 sentinel 分支测试，先证明当前实现还没有消费 `contest/ports` challenge lookup sentinel
- [ ] 为 `SubmissionRegistrationRepository` 补 challenge lookup not-found 映射测试，先证明 adapter 还没有承接这两个错误
- [ ] 跑最小测试，确认红灯来自目标 contract 缺失

验证：

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'SubmissionService.*Challenge' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'SubmissionRegistrationRepository' -count=1 -timeout 5m`

Review focus：

- service 测试是否真正在约束 submission sentinel，而不是继续借 `gorm` sentinel 过关
- adapter 测试是否只覆盖 challenge lookup not-found 映射，不夹带事务或写入逻辑

## Task 2: 实现 sentinel 映射与 submission 校验收口

**Files:**
- Modify: `code/backend/internal/module/contest/ports/submission.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_submit_validation.go`
- Modify: `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

- [ ] 在 `contest/ports/submission.go` 增加 submission challenge lookup not-found sentinel
- [ ] 让 `SubmissionRegistrationRepository` 承接 `FindContestChallenge` / `FindChallengeByID` 的 not-found 映射
- [ ] 让 `validateContestSubmission` 改成只看 sentinel，并继续映射为既有 `errcode`
- [ ] 保持 runtime wiring 仍然只注入同一个 adapter，不引入第二套 submission lookup 组合

验证：

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'SubmissionService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'Submission' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`

Review focus：

- `submission_submit_validation.go` 是否已经完全去掉 `gorm` concrete
- adapter 是否保持窄，只承接 challenge lookup not-found 映射

## Risks

- `FindChallengeByID` 同时服务校验和计分路径；这次只能改变 not-found 表达方式，不能把 scoring 逻辑一并改写
- 如果测试 helper 仍绕过 adapter 直接注入 raw repository，会导致代码和 runtime wiring 行为不一致
- 因用户明确禁止触碰 allowlist 与 shared docs，本 slice 只能交付代码 contract 收口；leader 后续如果需要清 shared evidence，需要单独处理

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/commands -run 'SubmissionService' -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -run 'Submission' -count=1 -timeout 300s`
3. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`

## Architecture-Fit Evaluation

- owner 明确：submission application 只保留业务校验和 `errcode` 映射；底层 not-found concrete 由 `ports + infrastructure adapter` 收口
- reuse point 明确：直接扩展现有 `SubmissionRegistrationRepository`，不再新建第二个 submission challenge adapter
- 结构收敛明确：本 slice 只处理 submission challenge lookup；AWD 和 shared allowlist 继续留在未触达边界之外
