# Challenge Topology Not-Found Contract Phase 5 Slice 29 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `challenge/application/commands/topology_service.go` 与 `challenge/application/queries/topology_service.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持 topology / template 查询与保存路径现有的 not-found 业务语义不变。

**Architecture:** `challenge` 新增一层窄 topology service adapter，把 raw topology repository、template repository、package revision repository 的 `gorm.ErrRecordNotFound` 收口成模块内 sentinel；topology command/query service 只消费 `challenge/ports` 错误契约，runtime builder 负责注入 adapter。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `challenge/application/commands/topology_service.go -> gorm.io/gorm`
- 删除 `challenge/application/queries/topology_service.go -> gorm.io/gorm`
- 保持以下语义不变：
  - challenge lookup not-found -> `errcode.ErrChallengeNotFound`
  - challenge topology / environment template not-found -> `errcode.ErrNotFound`
  - package revision detail not-found -> query 层忽略 package files，不把请求打成错误

## Non-goals

- 不修改 raw `challenge/infrastructure/writeup_repository.go` 的 GORM not-found 返回语义
- 不处理 contest / practice / auth / ops / assessment 代码
- 不修改共享文件：
  - `code/backend/internal/module/architecture_allowlist_test.go`
  - `docs/design/backend-module-boundary-target.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `.harness/reuse-decisions/echarts-mount-gate.md`

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/challenge/application/commands/topology_service.go`
- `code/backend/internal/module/challenge/application/queries/topology_service.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Ownership Boundary

- `challenge/application/commands/topology_service.go`
  - 负责：拓扑保存、模板写路径编排与 errcode 映射
  - 不负责：知道底层 not-found 是否来自 GORM
- `challenge/application/queries/topology_service.go`
  - 负责：拓扑详情、模板查询、package files 补充与 errcode 映射
  - 不负责：知道底层 not-found 是否来自 GORM
- `challenge/infrastructure/topology_service_repository.go`
  - 负责：把 raw topology/template/package-revision lookup 的 not-found 映射成 `challenge/ports` sentinel
  - 不负责：决定 application 最终返回哪个 errcode，或决定 package revision 缺失时是否忽略
- `challenge/runtime/module.go`
  - 负责：给 topology command/query service 注入 adapter
  - 不负责：把 raw GORM concrete 重新带回 topology service

## Change Surface

- Add: `.harness/reuse-decisions/challenge-topology-not-found-contract-phase5-slice29.md`
- Add: `docs/plan/impl-plan/2026-05-13-challenge-topology-not-found-contract-phase5-slice29-implementation-plan.md`
- Add: `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- Add: `code/backend/internal/module/challenge/infrastructure/topology_service_repository_test.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/application/commands/topology_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/topology_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/topology_service_context_test.go`
- Modify: `code/backend/internal/module/challenge/application/queries/topology_service_test.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`

## Architecture-Fit Evaluation

- owner boundary 是否明确：是。raw repository 和 template repository 继续是数据 owner；新增 adapter 只负责 not-found contract 收口。
- reuse point 是否明确：是。`challenge/ports` sentinel 是 topology command/query 共享契约，runtime 注入是唯一 wiring 落点。
- 是否只修行为不收结构：否。这次会把 `gorm.ErrRecordNotFound` 从 topology application surface 移出，而不是只改测试断言。
- touched surface 是否带已知结构债：是。topology command/query 直接依赖 GORM concrete 就是当前 slice 目标，会在这次一并收口。

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/topology_service_context_test.go`
- Modify: `code/backend/internal/module/challenge/application/queries/topology_service_test.go`
- Add: `code/backend/internal/module/challenge/infrastructure/topology_service_repository_test.go`

- [ ] 为 topology command/query 补 sentinel 分支测试，先证明当前实现仍依赖 GORM sentinel
- [ ] 为 topology adapter 补 raw GORM not-found -> ports sentinel 的映射测试
- [ ] 跑最小测试，确认红灯来自目标行为缺失

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'TopologyService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/queries -run 'TopologyService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'Topology|Template' -count=1 -timeout 300s`

Review focus：

- command/query 测试是否真正在约束 ports sentinel，而不是继续借 GORM sentinel 过关
- adapter 测试是否只覆盖 not-found 映射，不夹带 topology 业务逻辑

## Task 2: 实现 adapter 与 wiring

**Files:**
- Add: `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/application/commands/topology_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/topology_service.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`

- [ ] 在 `challenge/ports` 增加 topology/template/package-revision 相关 not-found sentinel
- [ ] 新增 topology service adapter，把 raw repo/template repo/package revision repo 的 `gorm.ErrRecordNotFound` 映射成对应 sentinel
- [ ] 让 topology command/query service 改成只看 sentinel
- [ ] 在 runtime wiring 中给 topology command/query service 注入 adapter

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'TopologyService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/queries -run 'TopologyService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'Topology|Template' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s`

Review focus：

- topology command/query surface 是否已经完全去掉 `gorm` concrete
- adapter 是否保持窄，只承接 topology/template/package-revision lookup not-found 映射

## Task 3: 最终验证与主线程 handoff

**Files:**
- 无 shared allowlist / shared facts docs 改动；整理验证结果和 handoff

- [ ] 跑用户指定的四条最小相关测试
- [ ] 跑 harness 一致性检查，确认新增 reuse decision / impl plan 没有破坏工作流
- [ ] 记录主线程仍需删除的两条 allowlist

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'TopologyService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/queries -run 'TopologyService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'Topology|Template' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`

Review focus：

- 共享文件是否保持未改，方便主线程统一整合
- handoff 是否明确指出要删除的两条 allowlist
