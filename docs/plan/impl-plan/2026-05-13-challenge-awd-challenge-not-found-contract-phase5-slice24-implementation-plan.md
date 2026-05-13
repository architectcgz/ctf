# Challenge AWD Challenge Not-Found Contract Phase 5 Slice 24 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `challenge/application/commands/awd_challenge_service.go` 与 `challenge/application/queries/awd_challenge_service.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持 AWD challenge 不存在时继续返回 `errcode.ErrNotFound`。

**Architecture:** `challenge` 新增一条窄 AWD challenge repository adapter，把 raw repository 的 `gorm.ErrRecordNotFound` 收口成模块内 sentinel；AWD challenge command/query service 只依赖 `challenge/ports` 错误契约，runtime builder 负责注入 adapter。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `challenge/application/commands/awd_challenge_service.go -> gorm.io/gorm`
- 删除 `challenge/application/queries/awd_challenge_service.go -> gorm.io/gorm`
- 保持 AWD challenge command/query 在资源不存在时继续返回 `errcode.ErrNotFound`

## Non-goals

- 不处理 `awd_challenge_import_service.go`、`awd_challenge_command_facade.go` 或其他 challenge GORM concrete
- 不修改 raw AWD challenge repository 的全局 not-found 返回语义
- 不重排 AWD import、image build 或其他 challenge runtime wiring

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_service.go`
- `code/backend/internal/module/challenge/application/queries/awd_challenge_service.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Ownership Boundary

- `challenge/application/commands/awd_challenge_service.go`
  - 负责：AWD challenge 创建、更新、删除编排和 errcode 映射
  - 不负责：知道底层 AWD challenge not-found 是否来自 `gorm`
- `challenge/application/queries/awd_challenge_service.go`
  - 负责：AWD challenge 详情、列表查询和 errcode 映射
  - 不负责：知道底层 AWD challenge not-found 是否来自 `gorm`
- `challenge/infrastructure/awd_challenge_repository.go`
  - 负责：把 raw repository 的 not-found 映射成 `challenge/ports` sentinel
  - 不负责：决定上层返回哪个 `errcode`
- `challenge/runtime/module.go`
  - 负责：给 AWD challenge command/query service 注入 adapter
  - 不负责：把 raw GORM concrete 重新带回 AWD challenge services

## Change Surface

- Add: `.harness/reuse-decisions/challenge-awd-challenge-not-found-contract-phase5-slice24.md`
- Add: `docs/plan/impl-plan/2026-05-13-challenge-awd-challenge-not-found-contract-phase5-slice24-implementation-plan.md`
- Add: `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- Add: `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository_test.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/queries/awd_challenge_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/awd_challenge_service_test.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/queries/awd_challenge_service_test.go`
- Add: `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository_test.go`

- [ ] 为 AWD challenge command/query service 补 sentinel 分支测试，先证明当前实现仍依赖 GORM sentinel
- [ ] 为 AWD challenge adapter 补 GORM not-found 映射测试，先证明 adapter 尚不存在
- [ ] 跑最小测试，确认红灯来自目标行为缺失

验证：

- `cd code/backend && go test ./internal/module/challenge/application/commands -run 'AWDChallengeServiceTreatsAWDChallengeNotFoundAsNotFound' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/challenge/application/queries -run 'AWDChallengeQueryServiceTreatsAWDChallengeNotFoundAsNotFound' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/challenge/infrastructure -run 'AWDChallengeRepository' -count=1 -timeout 5m`

Review focus：

- service 测试是否真正在约束模块内 sentinel，而不是继续借 GORM sentinel 过关
- adapter 测试是否只覆盖 not-found 映射，不夹带 AWD challenge 业务逻辑

## Task 2: 实现 adapter 与 wiring

**Files:**
- Add: `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/awd_challenge_service.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`

- [ ] 在 `challenge/ports` 增加 AWD challenge not-found sentinel
- [ ] 新增 AWD challenge adapter，把 raw repository 的 `gorm.ErrRecordNotFound` 映射成模块内 sentinel
- [ ] 让 command/query service 改成只看 sentinel
- [ ] 在 runtime wiring 中给 AWD challenge command/query service 注入 adapter

验证：

- `cd code/backend && go test ./internal/module/challenge/application/commands -run 'AWDChallengeService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/challenge/application/queries -run 'AWDChallengeQueryService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/challenge/infrastructure -run 'AWDChallengeRepository' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 5m`

Review focus：

- AWD challenge command/query surface 是否已经完全去掉 `gorm` concrete
- adapter 是否保持窄，只承接 AWD challenge lookup not-found 映射

## Task 3: 删除 allowlist 并同步文档

**Files:**
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

- [ ] 删除本 slice 实际收口的 allowlist
- [ ] 更新 phase5 当前状态，记录 AWD challenge not-found contract 已完成
- [ ] 跑架构 / 文档 / workflow 完整性检查

验证：

- `cd code/backend && go test ./internal/module -run 'Test(ApplicationConcreteDependencyAllowlistIsCurrent|ModuleDependencyAllowlistIsCurrent)' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `timeout 600 bash scripts/check-workflow-complete.sh`
- `git diff --check`

Review focus：

- 只删除已经真实收口的 allowlist
- 文档是否准确描述“AWD challenge adapter + sentinel + runtime wiring”的 owner 分工
