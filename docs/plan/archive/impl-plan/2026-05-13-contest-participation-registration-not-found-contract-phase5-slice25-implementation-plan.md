# Contest Participation Registration Not-Found Contract Phase 5 Slice 25 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `contest/application/commands/participation_register_commands.go`、`contest/application/commands/participation_review_commands.go`、`contest/application/commands/submission_validation.go`、`contest/application/queries/participation_progress_query.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持报名不存在、用户未组队和未报名 fallback 的业务行为不变。

**Architecture:** `contest` 新增 registration lookup adapter 与 team finder adapter，把 raw repository 的 `gorm.ErrRecordNotFound` 收口成模块内 sentinel；participation / submission application 只依赖 `contest/ports` 错误契约，runtime builder 负责注入 adapter。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `contest/application/commands/participation_register_commands.go -> gorm.io/gorm`
- 删除 `contest/application/commands/participation_review_commands.go -> gorm.io/gorm`
- 删除 `contest/application/commands/submission_validation.go -> gorm.io/gorm`
- 删除 `contest/application/queries/participation_progress_query.go -> gorm.io/gorm`

## Non-goals

- 不处理 `team_support.go`、`team_join_commands.go`、`team_info_query.go` 等 team repository 其他 not-found concrete
- 不处理 `submission_submit_validation.go`、`contest_awd_service_service.go`、AWD jobs/query 里的其他 GORM sentinel
- 不修改 raw `ParticipationRepository`、`SubmissionRepository`、`TeamRepository` 的全局 not-found 返回语义
- 不重排 contest module 的其他 Redis / HTTP wiring

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/contest/application/commands/participation_register_commands.go`
- `code/backend/internal/module/contest/application/commands/participation_review_commands.go`
- `code/backend/internal/module/contest/application/commands/submission_validation.go`
- `code/backend/internal/module/contest/application/queries/participation_progress_query.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/participation.go`
- `code/backend/internal/module/contest/ports/team.go`

## Ownership Boundary

- `contest/application/commands/participation_register_commands.go`
  - 负责：报名创建/重提编排，决定未报名、未组队时的业务 fallback
  - 不负责：知道 registration / team lookup not-found 是否来自 GORM
- `contest/application/commands/participation_review_commands.go`
  - 负责：报名审核编排和公开 errcode 映射
  - 不负责：知道 registration lookup not-found 是否来自 GORM
- `contest/application/commands/submission_validation.go`
  - 负责：提交前用户报名/队伍归属判断和 `ErrNotRegistered` 映射
  - 不负责：知道 registration / team lookup not-found 是否来自 GORM
- `contest/application/queries/participation_progress_query.go`
  - 负责：我的进度查询和 team fallback 判定
  - 不负责：知道 registration / team lookup not-found 是否来自 GORM
- `contest/infrastructure/participation_registration_repository.go`
  - 负责：把 participation raw repository 的 registration lookup not-found 映射成模块内 sentinel
  - 不负责：决定上层返回哪个 errcode 或 fallback
- `contest/infrastructure/submission_registration_repository.go`
  - 负责：把 submission raw repository 的 registration lookup not-found 映射成模块内 sentinel
  - 不负责：承载 scoring transaction、contest challenge lookup 的业务决策
- `contest/infrastructure/team_finder_repository.go`
  - 负责：把 `FindUserTeamInContest` 的 not-found 映射成模块内 sentinel
  - 不负责：承接 team detail / team membership / registration binding 的其他语义
- `contest/runtime/module.go`
  - 负责：给 participation / submission service 注入 adapter
  - 不负责：把 raw GORM concrete 重新带回 application surface

## Change Surface

- Add: `.harness/reuse-decisions/contest-participation-registration-not-found-contract-phase5-slice25.md`
- Add: `docs/plan/impl-plan/2026-05-13-contest-participation-registration-not-found-contract-phase5-slice25-implementation-plan.md`
- Add: `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- Add: `code/backend/internal/module/contest/infrastructure/submission_registration_repository_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/team_finder_repository.go`
- Add: `code/backend/internal/module/contest/infrastructure/team_finder_repository_test.go`
- Modify: `code/backend/internal/module/contest/infrastructure/participation_registration_repository.go`
- Modify: `code/backend/internal/module/contest/infrastructure/participation_registration_repository_test.go`
- Modify: `code/backend/internal/module/contest/ports/participation.go`
- Modify: `code/backend/internal/module/contest/ports/team.go`
- Modify: `code/backend/internal/module/contest/application/commands/participation_register_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/participation_review_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_validation.go`
- Modify: `code/backend/internal/module/contest/application/commands/participation_service.go`
- Modify: `code/backend/internal/module/contest/application/queries/participation_progress_query.go`
- Modify: `code/backend/internal/module/contest/application/queries/participation_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/participation_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/contest/application/commands/participation_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_service_test.go`
- Modify: `code/backend/internal/module/contest/infrastructure/participation_registration_repository_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/submission_registration_repository_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/team_finder_repository_test.go`

- [ ] 为 `ReviewRegistration`、`resolveTeamID`、`resolveUserTeamID` 补 sentinel 分支测试，先证明当前实现仍依赖 GORM sentinel
- [ ] 为 registration / team finder adapter 补 GORM not-found 映射测试，先证明 adapter 尚不存在
- [ ] 跑最小测试，确认红灯来自目标行为缺失

验证：

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'ParticipationServiceTreatsRegistrationNotFoundAsContestRegistrationNotFound|SubmissionServiceResolveTeamID' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/application/queries -run 'ParticipationServiceResolveUserTeamID' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'ParticipationRegistrationRepository|SubmissionRegistrationRepository|TeamFinderRepository' -count=1 -timeout 5m`

Review focus：

- application 测试是否真正在约束模块内 sentinel，而不是继续借 GORM sentinel 过关
- adapter 测试是否只覆盖 not-found 映射，不夹带报名或队伍业务逻辑

## Task 2: 实现 adapter 与 wiring

**Files:**
- Modify: `code/backend/internal/module/contest/infrastructure/participation_registration_repository.go`
- Add: `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- Add: `code/backend/internal/module/contest/infrastructure/team_finder_repository.go`
- Modify: `code/backend/internal/module/contest/ports/participation.go`
- Modify: `code/backend/internal/module/contest/ports/team.go`
- Modify: `code/backend/internal/module/contest/application/commands/participation_register_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/participation_review_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_validation.go`
- Modify: `code/backend/internal/module/contest/application/commands/participation_service.go`
- Modify: `code/backend/internal/module/contest/application/queries/participation_progress_query.go`
- Modify: `code/backend/internal/module/contest/application/queries/participation_service.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

- [ ] 在 `contest/ports` 增加 registration/team finder not-found sentinel
- [ ] 让 participation registration adapter 统一承接 command/query service 使用的 raw participation repo
- [ ] 新增 submission registration adapter 与 team finder adapter，统一把 `gorm.ErrRecordNotFound` 收口成模块内 sentinel
- [ ] 让 participation / submission application 改成只看 sentinel
- [ ] 在 runtime wiring 中给 participation / submission service 注入 adapter

验证：

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'ParticipationService|SubmissionService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/application/queries -run 'ParticipationService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'ParticipationRegistrationRepository|SubmissionRegistrationRepository|TeamFinderRepository' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`

Review focus：

- participation / submission application surface 是否已经完全去掉目标 GORM concrete
- team finder adapter 是否保持窄，只承接 `FindUserTeamInContest` 的 not-found 语义

## Task 3: 删除 allowlist 并同步文档

**Files:**
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

- [ ] 删除本 slice 实际收口的 allowlist
- [ ] 更新 phase5 当前状态，记录 contest participation/submission registration 与 team finder not-found contract 已完成
- [ ] 跑架构 / 文档 / workflow 完整性检查

验证：

- `cd code/backend && go test ./internal/module -run 'Test(ApplicationConcreteDependencyAllowlistIsCurrent|ModuleDependencyAllowlistIsCurrent)' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `timeout 600 bash scripts/check-workflow-complete.sh`
- `git diff --check`

Review focus：

- 只删除已经真实收口的 allowlist
- 文档是否准确描述 “registration/team finder sentinel + adapter + runtime wiring + application errcode mapping” 的 owner 分工
