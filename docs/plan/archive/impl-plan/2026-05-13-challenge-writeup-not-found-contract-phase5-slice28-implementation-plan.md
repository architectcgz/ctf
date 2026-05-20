# Challenge Writeup Not-Found Contract Phase 5 Slice 28 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `challenge/application/commands/writeup_service.go` 与 `challenge/application/queries/writeup_service.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持各 writeup use case 现有的 not-found 业务语义不变。

**Architecture:** `challenge` 新增一层窄 writeup adapter，把 raw `Repository` 的不同 writeup lookup not-found 统一收口成模块内 sentinel；writeup command/query service 只消费 `challenge/ports` 错误契约，runtime builder 负责注入 adapter。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `challenge/application/commands/writeup_service.go -> gorm.io/gorm`
- 删除 `challenge/application/queries/writeup_service.go -> gorm.io/gorm`
- 保持 writeup command/query 现有的 errcode / nil fallback 语义

## Non-goals

- 不修改 raw `challenge/infrastructure/writeup_repository.go` 的 GORM not-found 返回语义
- 不处理 `challenge` 模块其他残留的 GORM concrete
- 不修改共享文件：
  - `code/backend/internal/module/architecture_allowlist_test.go`
  - `docs/design/backend-module-boundary-target.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/challenge/application/commands/writeup_service.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Ownership Boundary

- `challenge/application/commands/writeup_service.go`
  - 负责：官方题解 / 社区题解写路径编排与 errcode 映射
  - 不负责：知道底层 not-found 是否来自 GORM
- `challenge/application/queries/writeup_service.go`
  - 负责：写路径查询、教师查看和 errcode / nil fallback 映射
  - 不负责：知道底层 not-found 是否来自 GORM
- `challenge/infrastructure/writeup_service_repository.go`
  - 负责：把 raw writeup repository 的不同 lookup not-found 映射成 `challenge/ports` sentinel
  - 不负责：决定 application 最终返回哪个 errcode 或 nil fallback
- `challenge/runtime/module.go`
  - 负责：给 writeup command/query service 注入 adapter
  - 不负责：把 raw GORM concrete 重新带回 writeup service

## Change Surface

- Add: `.harness/reuse-decisions/challenge-writeup-not-found-contract-phase5-slice28.md`
- Add: `docs/plan/impl-plan/2026-05-13-challenge-writeup-not-found-contract-phase5-slice28-implementation-plan.md`
- Add: `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`
- Add: `code/backend/internal/module/challenge/infrastructure/writeup_service_repository_test.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/application/commands/writeup_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/writeup_service_context_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/writeup_submission_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/writeup_topology_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/queries/writeup_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/writeup_service_test.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`

## Architecture-Fit Evaluation

- owner boundary 是否明确：是。raw repository 继续是数据 owner；新增 adapter 只收口 not-found contract。
- reuse point 是否明确：是。`challenge/ports` sentinel 是 writeup command/query 共享契约，runtime 注入是唯一 wiring 落点。
- 是否只修行为不收结构：否。这次会把 `gorm.ErrRecordNotFound` 从 writeup application surface 移出，而不是只改断言。
- touched surface 是否带已知结构债：是。writeup command/query 直接依赖 GORM concrete 就是当前 slice 目标，会在这次一并收口。

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/writeup_service_context_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/writeup_submission_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/writeup_topology_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/queries/writeup_service_test.go`
- Add: `code/backend/internal/module/challenge/infrastructure/writeup_service_repository_test.go`

- [x] 为 writeup command/query 补 sentinel 分支测试，先证明当前实现仍依赖 GORM sentinel
- [x] 为 writeup adapter 补 raw GORM not-found -> ports sentinel 的映射测试
- [x] 跑最小测试，确认红灯来自目标行为缺失

验证：

- `cd code/backend && go test ./internal/module/challenge/application/commands -run 'WriteupService' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/challenge/application/queries -run 'WriteupService' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/challenge/infrastructure -run 'WriteupRepository' -count=1 -timeout 300s`

Review focus：

- command/query 测试是否真正在约束 ports sentinel，而不是继续借 GORM sentinel 过关
- adapter 测试是否只覆盖 not-found 映射，不夹带题解业务逻辑

## Task 2: 实现 adapter 与 wiring

**Files:**
- Add: `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/application/commands/writeup_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/writeup_service.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`

- [x] 在 `challenge/ports` 增加 writeup 相关 not-found sentinel
- [x] 新增 writeup adapter，把 raw repo 的各类 `gorm.ErrRecordNotFound` 映射成对应 sentinel
- [x] 让 writeup command/query service 改成只看 sentinel
- [x] 在 runtime wiring 中给 writeup command/query service 注入 adapter

验证：

- `cd code/backend && go test ./internal/module/challenge/application/commands -run 'WriteupService' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/challenge/application/queries -run 'WriteupService' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/challenge/infrastructure -run 'WriteupRepository' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s`

Review focus：

- writeup command/query surface 是否已经完全去掉 `gorm` concrete
- adapter 是否保持窄，只承接 writeup lookup not-found 映射

## Task 3: 最终验证与主线程 handoff

**Files:**
- 无新增代码文件；整理验证结果和 shared allowlist handoff

- [x] 跑用户指定的四条最小相关测试
- [x] 确认本次 slice 未新增共享 allowlist/docs 改动
- [x] 记录主线程仍需删除的 allowlist 行

验证：

- `cd code/backend && go test ./internal/module/challenge/application/commands -run 'WriteupService' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/challenge/application/queries -run 'WriteupService' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/challenge/infrastructure -run 'WriteupRepository' -count=1 -timeout 300s`
- `cd code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s`

Review focus：

- 共享文件是否保持未改，方便主线程统一整合
- handoff 是否明确指出要删除的两条 allowlist
