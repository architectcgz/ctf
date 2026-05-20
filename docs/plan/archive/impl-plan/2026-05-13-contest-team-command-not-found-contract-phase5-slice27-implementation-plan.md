# Contest Team Command Not-Found Contract Phase 5 Slice 27 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `contest/application/commands/team_support.go`、`team_join_commands.go`、`team_leave_commands.go`、`team_captain_manage_commands.go`、`team_create_retry_support.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持既有业务返回：team not found -> `errcode.ErrTeamNotFound`，registration missing -> `errcode.ErrNotRegistered`，user current team missing -> `nil` fallback / 继续创建或加入。

**Architecture:** 在 `contest/infrastructure` 增加一个 command-only team adapter，把 raw `TeamRepository` 的 lookup 与 registration binding not-found 翻译成 `contest/ports` sentinel；command `TeamService` 只消费 ports/domain sentinel，runtime builder 单独给 command service 注入这个 adapter，query service 保持现有 `team_query_adapter`。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `contest/application/commands/team_support.go -> gorm.io/gorm`
- 删除 `contest/application/commands/team_join_commands.go -> gorm.io/gorm`
- 删除 `contest/application/commands/team_leave_commands.go -> gorm.io/gorm`
- 删除 `contest/application/commands/team_captain_manage_commands.go -> gorm.io/gorm`
- 删除 `contest/application/commands/team_create_retry_support.go -> gorm.io/gorm`

## Non-goals

- 不修改 raw `TeamRepository` 的全局 not-found 返回语义
- 不处理 AWD、challenge 或 shared docs / allowlist 文件
- 不改 query `TeamService` 的 `team_query_adapter` 路径
- 不扩展为 team repository 大拆分或 command/query 合并重构

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/contest/application/commands/team_service.go`
- `code/backend/internal/module/contest/application/commands/team_support.go`
- `code/backend/internal/module/contest/application/commands/team_join_commands.go`
- `code/backend/internal/module/contest/application/commands/team_leave_commands.go`
- `code/backend/internal/module/contest/application/commands/team_captain_manage_commands.go`
- `code/backend/internal/module/contest/application/commands/team_create_retry_support.go`
- `code/backend/internal/module/contest/application/commands/team_create_commands.go`
- `code/backend/internal/module/contest/infrastructure/team_query_adapter.go`
- `code/backend/internal/module/contest/infrastructure/team_finder_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/team.go`
- `code/backend/internal/module/contest/ports/participation.go`

## Ownership Boundary

- `contest/application/commands/team_*.go`
  - 负责：命令语义、业务 errcode 映射和 fallback 决策
  - 不负责：知道 raw repository 的 not-found 是否来自 GORM
- `contest/infrastructure/team_command_adapter.go`
  - 负责：把 team command surface 需要的 not-found 收口成 ports sentinel
  - 不负责：改变 raw repository 的事务、唯一约束或写路径实现
- `contest/runtime/module.go`
  - 负责：给 command `TeamService` 注入 command adapter
  - 不负责：改 query `TeamService` 现有 wiring

## Change Surface

- Add: `.harness/reuse-decisions/contest-team-command-not-found-contract-phase5-slice27.md`
- Add: `docs/plan/impl-plan/2026-05-13-contest-team-command-not-found-contract-phase5-slice27-implementation-plan.md`
- Add: `code/backend/internal/module/contest/application/commands/team_error_contract_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/team_command_adapter.go`
- Add: `code/backend/internal/module/contest/infrastructure/team_command_adapter_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_join_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_leave_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_captain_manage_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_create_retry_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_create_commands.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

## Task 1: 先写失败测试

**Files:**

- Add: `code/backend/internal/module/contest/application/commands/team_error_contract_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/team_command_adapter_test.go`

- [ ] 为 `CreateTeam` / `JoinTeam` / `LeaveTeam` / `DismissTeam` / `KickMember` 补 sentinel 分支测试，证明当前 command 实现还直接依赖 GORM concrete 或错误地吞掉非 sentinel 错误
- [ ] 为 `TeamCommandAdapter` 补 not-found 映射测试，证明 command adapter 还不存在
- [ ] 跑最小测试，确认红灯来自目标行为缺失

验证：

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'TeamService' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'TeamCommandAdapter' -count=1 -timeout 300s`

Review focus：

- command 测试是否真正在约束 ports sentinel，而不是继续借 `gorm.ErrRecordNotFound` 过关
- adapter 测试是否只覆盖 not-found contract，不夹带仓储实现细节

## Task 2: 实现 command adapter 与 wiring

**Files:**

- Add: `code/backend/internal/module/contest/infrastructure/team_command_adapter.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_join_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_leave_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_captain_manage_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_create_retry_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/team_create_commands.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

- [ ] 新增 command-only team adapter，收口 `FindByID` / `FindContestRegistration` / `FindUserTeamInContest` 的 not-found
- [ ] 让 command adapter 同时把 `CreateWithMember` / `AddMemberWithLock` 中 registration binding 暴露出的 not-found 收口成 `ErrContestParticipationRegistrationNotFound`
- [ ] 让 `TeamService` 只分支 ports/domain sentinel，非 sentinel 错误走 internal
- [ ] runtime wiring 改成 command 使用 command adapter，query 继续使用 `team_query_adapter`

验证：

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'TeamService' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'TeamCommandAdapter' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`

Review focus：

- team command surface 是否完全去掉目标 GORM concrete
- command adapter 是否只承接当前 slice 需要的 not-found 翻译，不扩大模块边界

## Task 3: 复核 handoff 边界

**Files:**

- Modify: `docs/plan/impl-plan/2026-05-13-contest-team-command-not-found-contract-phase5-slice27-implementation-plan.md`

- [ ] 记录本 slice 已收口的 allowlist 行，留给主线程统一删除共享文件
- [ ] 确认没有误碰 AWD、challenge、shared docs

验证：

- `git diff --check`

Review focus：

- handoff 是否只包含主线程还需要统一整合的 allowlist 删除
