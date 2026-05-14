# Contest AWD Command Lookup Contract Phase 5 Slice 37 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `contest/application/commands` 中这 7 个 AWDService command 文件对 `gorm.ErrRecordNotFound` 的直接依赖，并保持 AWD command 既有业务语义不变。

**Architecture:** 在 `contest/infrastructure` 增加 command-only AWD repository adapter 与 round-manager adapter，把 raw repo / round manager 暴露出来的 lookup not-found 映射成 `contest/ports` sentinel；`contest/application/commands` 只消费模块语义错误；runtime 只给 `AWDService` command 路径注入 adapter，query 与 jobs 继续保留现有 wiring。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `contest/application/commands/awd_attack_log_commands.go -> gorm.io/gorm`
- 删除 `contest/application/commands/awd_current_round_active_support.go -> gorm.io/gorm`
- 删除 `contest/application/commands/awd_current_round_fallback_support.go -> gorm.io/gorm`
- 删除 `contest/application/commands/awd_flag_support.go -> gorm.io/gorm`
- 删除 `contest/application/commands/awd_resource_validation_support.go -> gorm.io/gorm`
- 删除 `contest/application/commands/awd_team_validation_support.go -> gorm.io/gorm`
- 删除 `contest/application/commands/awd_validation_support.go -> gorm.io/gorm`

## Non-goals

- 不修改 `code/backend/internal/module/architecture_allowlist_test.go`
- 不修改 `docs/architecture/backend/07-modular-monolith-refactor.md`
- 不修改 `docs/design/backend-module-boundary-target.md`
- 不修改 `contest/application/jobs/*` 那批 updater/runtime 文件
- 不扩展到 `challenge_add_commands.go` 或 `contest_awd_service_service.go`
- 不扩展到 http handler、query service 或 teaching 脏改
- 不修改 raw `AWDRepository` 的全局 not-found 语义

## Inputs

- `.harness/reuse-decisions/contest-awd-command-lookup-contract-phase5-slice37.md`
- `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_current_round_active_support.go`
- `code/backend/internal/module/contest/application/commands/awd_current_round_fallback_support.go`
- `code/backend/internal/module/contest/application/commands/awd_flag_support.go`
- `code/backend/internal/module/contest/application/commands/awd_resource_validation_support.go`
- `code/backend/internal/module/contest/application/commands/awd_team_validation_support.go`
- `code/backend/internal/module/contest/application/commands/awd_validation_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service.go`
- `code/backend/internal/module/contest/infrastructure/team_command_adapter.go`
- `code/backend/internal/module/contest/infrastructure/awd_query_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/ports/team.go`
- `code/backend/internal/module/contest/ports/participation.go`

## Ownership Boundary

- `contest/application/commands/awd_*.go`
  - 负责：AWD command 业务语义、fallback 规则、errcode 映射
  - 不负责：知道底层 lookup not-found 是否来自 GORM
- `contest/ports/awd.go`
  - 负责：定义本 slice 需要的 AWD command-side sentinel
  - 不负责：决定 HTTP 返回码
- `contest/infrastructure/awd_command_repository.go`
  - 负责：把 AWD command path 上的 raw lookup / tx not-found 收口成模块 sentinel
  - 不负责：改变 raw repo 的事务、查询和写入语义
- `contest/infrastructure/awd_round_manager_adapter.go`
  - 负责：把 round manager 暴露出的 round not-found 收口成模块 sentinel
  - 不负责：改变 jobs/updater 内部逻辑
- `contest/runtime/module.go`
  - 负责：只给 command `AWDService` 注入 adapter
  - 不负责：改动 jobs / query / http 其他 wiring

## Change Surface

- Add: `.harness/reuse-decisions/contest-awd-command-lookup-contract-phase5-slice37.md`
- Add: `docs/plan/impl-plan/2026-05-13-contest-awd-command-lookup-contract-phase5-slice37-implementation-plan.md`
- Add: `code/backend/internal/module/contest/infrastructure/awd_command_repository.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_command_repository_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_round_manager_adapter.go`
- Add: `code/backend/internal/module/contest/application/commands/awd_error_contract_test.go`
- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_current_round_active_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_current_round_fallback_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_flag_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_resource_validation_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_team_validation_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_validation_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

## Task 1: 先写失败测试

**Files:**

- Add: `code/backend/internal/module/contest/application/commands/awd_error_contract_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_command_repository_test.go`

- [ ] 为 AWD command-side round / registration / team / challenge / service sentinel 分支补齐 application contract 测试
- [ ] 为 repository adapter 与 round-manager adapter 补 raw GORM not-found -> contest sentinel 测试
- [ ] 跑最小测试，确认红灯来自 command 代码尚未消费模块 sentinel

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/commands -run 'AWDService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -run 'AWD(CommandRepository|RoundManagerAdapter)' -count=1 -timeout 300s`

Review focus：

- contract 测试是否在约束模块 sentinel，而不是继续借 `gorm.ErrRecordNotFound` 过关
- adapter 测试是否只覆盖 not-found contract，不扩展 raw repo / jobs 行为

## Task 2: 实现 command adapter、ports sentinel 与 wiring

**Files:**

- Add: `code/backend/internal/module/contest/infrastructure/awd_command_repository.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_round_manager_adapter.go`
- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_current_round_active_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_current_round_fallback_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_flag_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_resource_validation_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_team_validation_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_validation_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

- [ ] 在 `contest/ports/awd.go` 增加本 slice 需要的 command-side sentinel
- [ ] 新增 `AWDCommandRepository`，收口 `FindRoundByContestAndID`、`FindRoundByNumber`、`FindRunningRound`、`FindRegistration`、`FindContestTeamByMember`、`FindChallengeByID`、`FindContestAWDServiceByContestAndID` 与 attack-log tx not-found
- [ ] 新增 `AWDRoundManagerAdapter`，收口 `EnsureActiveRoundMaterialized` 的 round not-found
- [ ] 让 AWD command 只消费 contest sentinel，删除 `gorm` import
- [ ] runtime 只在 `AWDService` command 路径注入 adapter；jobs/query 继续保持 raw wiring

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/commands -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`

Review focus：

- 7 个目标 command 文件是否已经完全脱离 `gorm` concrete
- runtime 是否只改 command wiring，没有误碰 jobs/http/query surface
- fallback 语义是否保持不变：current round 缺失、team 缺失、previous round 缺失、challenge 缺失、service 缺失仍返回既有 errcode / 空行为

## Risks

- `AWDService` 同时包含 round lookup、registration/team fallback、attack-log transaction；如果 adapter 映射过宽，可能把真实内部错误误吞成 404
- command helper 测试如果继续直接注入 raw repo，会把旧的 `gorm.ErrRecordNotFound` 重新带回 command service
- 本 slice 明确不处理 `challenge_add_commands.go`、`contest_awd_service_service.go`、`contest/application/jobs/*` 和 shared cleanup；最终 allowlist / architecture facts 仍需 leader 统一收口

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/commands -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -count=1 -timeout 300s`
3. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`

## Architecture-Fit Evaluation

- owner 明确：not-found concrete 由 `contest/infrastructure + contest/ports` 收口，command application 只决定业务 errcode
- reuse point 明确：复用 phase5 已验证的 “sentinel 在 ports、映射在 infrastructure、runtime 定向注入” 模式，不改 raw repo 全局语义
- 结构收敛明确：只处理 AWDService command-side 这 7 条 allowlist，不扩到 challenge command、jobs/http/shared cleanup
