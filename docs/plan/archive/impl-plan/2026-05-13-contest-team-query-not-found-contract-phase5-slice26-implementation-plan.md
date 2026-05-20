# Contest Team Query Not-Found Contract Phase 5 Slice 26 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `contest/application/queries/team_info_query.go` 与 `contest/application/queries/team_list_query.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持 team detail not-found 继续返回 `errcode.ErrTeamNotFound`，用户未加入队伍时 `GetMyTeam` 继续返回 `nil`。

**Architecture:** `contest` 新增一个 query-only team adapter，把 raw `TeamRepository` 的 `FindByID` / `FindUserTeamInContest` not-found 语义收口成模块内 sentinel；team query service 只依赖 `contest/ports` 错误契约，runtime builder 负责把 adapter 注入 query wiring。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `contest/application/queries/team_info_query.go -> gorm.io/gorm`
- 删除 `contest/application/queries/team_list_query.go -> gorm.io/gorm`

## Non-goals

- 不处理 `team_support.go`、`team_join_commands.go`、`team_leave_commands.go`、`team_captain_manage_commands.go` 等 command surface
- 不修改 raw `TeamRepository` 的全局 not-found 返回语义
- 不把 team membership write、registration binding 或 invite code 逻辑拉进这一刀

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/contest/application/queries/team_info_query.go`
- `code/backend/internal/module/contest/application/queries/team_list_query.go`
- `code/backend/internal/module/contest/application/queries/team_service.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/team.go`

## Ownership Boundary

- `contest/application/queries/team_info_query.go`
  - 负责：队伍详情读取、成员拼装和公开 errcode 映射
  - 不负责：知道 `FindByID` not-found 是否来自 GORM
- `contest/application/queries/team_list_query.go`
  - 负责：队伍列表读取、我的队伍 fallback 和结果拼装
  - 不负责：知道 `FindUserTeamInContest` not-found 是否来自 GORM
- `contest/infrastructure/team_query_adapter.go`
  - 负责：把 `FindByID` / `FindUserTeamInContest` 的 not-found 收口成模块内 sentinel，并 passthrough 其他 team query 方法
  - 不负责：承接 team command、membership write 或 registration binding 语义
- `contest/runtime/module.go`
  - 负责：给 query `TeamService` 注入 team query adapter
  - 不负责：把 raw GORM concrete 重新带回 team query surface

## Change Surface

- Add: `.harness/reuse-decisions/contest-team-query-not-found-contract-phase5-slice26.md`
- Add: `docs/plan/impl-plan/2026-05-13-contest-team-query-not-found-contract-phase5-slice26-implementation-plan.md`
- Add: `code/backend/internal/module/contest/infrastructure/team_query_adapter.go`
- Add: `code/backend/internal/module/contest/infrastructure/team_query_adapter_test.go`
- Modify: `code/backend/internal/module/contest/ports/team.go`
- Modify: `code/backend/internal/module/contest/application/queries/team_info_query.go`
- Modify: `code/backend/internal/module/contest/application/queries/team_list_query.go`
- Modify: `code/backend/internal/module/contest/application/queries/team_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/contest/application/queries/team_service_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/team_query_adapter_test.go`

- [ ] 为 `GetTeamInfo` / `GetMyTeam` 补 sentinel 分支测试，先证明当前实现仍依赖 GORM sentinel
- [ ] 为 team query adapter 补 GORM not-found 映射测试，先证明 adapter 尚不存在
- [ ] 跑最小测试，确认红灯来自目标行为缺失

验证：

- `cd code/backend && go test ./internal/module/contest/application/queries -run 'TeamService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'TeamQueryAdapter' -count=1 -timeout 5m`

Review focus：

- query 测试是否真正在约束模块内 sentinel，而不是继续借 GORM sentinel 过关
- adapter 是否只承接 query lookup not-found 语义，不夹带 write 逻辑

## Task 2: 实现 adapter 与 wiring

**Files:**
- Add: `code/backend/internal/module/contest/infrastructure/team_query_adapter.go`
- Modify: `code/backend/internal/module/contest/ports/team.go`
- Modify: `code/backend/internal/module/contest/application/queries/team_info_query.go`
- Modify: `code/backend/internal/module/contest/application/queries/team_list_query.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`

- [ ] 在 `contest/ports` 增加 team detail not-found sentinel
- [ ] 新增 query-only team adapter，统一把 raw `TeamRepository` 的 `FindByID` / `FindUserTeamInContest` not-found 收口成模块内 sentinel
- [ ] 让 team query service 改成只看 sentinel
- [ ] 在 runtime wiring 中只给 query `TeamService` 注入 adapter，command service 继续保留 raw repo

验证：

- `cd code/backend && go test ./internal/module/contest/application/queries -run 'TeamService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/infrastructure -run 'TeamQueryAdapter' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`

Review focus：

- team query surface 是否已经完全去掉目标 GORM concrete
- runtime wiring 是否只影响 query service，不扩大到 team command surface

## Task 3: 删除 allowlist 并同步文档

**Files:**
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

- [ ] 删除本 slice 实际收口的 allowlist
- [ ] 更新 phase5 当前状态，记录 team query not-found contract 已完成
- [ ] 跑架构 / 文档 / workflow 完整性检查

验证：

- `cd code/backend && go test ./internal/module -run 'Test(ApplicationConcreteDependencyAllowlistIsCurrent|ModuleDependencyAllowlistIsCurrent)' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `timeout 600 bash scripts/check-workflow-complete.sh`
- `git diff --check`

Review focus：

- 只删除已经真实收口的 allowlist
- 文档是否准确描述 “team query sentinel + adapter + runtime wiring” 的 owner 分工
