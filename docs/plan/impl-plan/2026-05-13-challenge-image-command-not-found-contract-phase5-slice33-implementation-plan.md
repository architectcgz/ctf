# Challenge Image Command Not-Found Contract Phase 5 Slice 33 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `challenge/application/commands/image_service.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持 image command 既有语义不变：重复镜像仍拦截、image not-found 仍映射到 `errcode.ErrImageNotFound`、同名镜像未找到时仍允许创建。

**Architecture:** 新增一层窄 `ImageCommandRepository` adapter，把 raw image repository 的 lookup not-found 收口成 `challenge/ports.ErrChallengeImageNotFound`；`ImageService` 只消费该 sentinel；runtime builder 只给 image command service 注入 adapter，query/build 继续使用现有注入面。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `challenge/application/commands/image_service.go -> gorm.io/gorm`
- 保持 `CreateImage` 在“同名镜像不存在”时继续创建
- 保持 `UpdateImage` / `DeleteImage` 的 image lookup not-found -> `errcode.ErrImageNotFound`

## Non-goals

- 不修改 `code/backend/internal/module/architecture_allowlist_test.go`
- 不修改 `docs/architecture/backend/07-modular-monolith-refactor.md`
- 不修改 `docs/design/backend-module-boundary-target.md`
- 不处理 image build transaction surface
- 不处理 challenge import / package revision / challenge core command 其他 concrete 依赖
- 不改 contest 模块代码

## Inputs

- `code/backend/internal/module/challenge/application/commands/image_service.go`
- `code/backend/internal/module/challenge/application/commands/image_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/image_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Ownership Boundary

- `challenge/application/commands/image_service.go`
  - 负责：image command 编排、已有 `errcode` 映射、异步删除触发
  - 不负责：知道 lookup not-found 是否来自 GORM
- `challenge/infrastructure/image_command_repository.go`
  - 负责：把 raw image repository 的 command lookup not-found 映射成模块 sentinel
  - 不负责：改变 create/update/delete 的写入语义
- `challenge/runtime/module.go`
  - 负责：只给 image command service 注入 adapter
  - 不负责：把 adapted 行为扩散到 image build/query 之外的调用方

## Change Surface

- Add: `.harness/reuse-decisions/challenge-image-command-not-found-contract-phase5-slice33.md`
- Add: `docs/plan/impl-plan/2026-05-13-challenge-image-command-not-found-contract-phase5-slice33-implementation-plan.md`
- Add: `code/backend/internal/module/challenge/infrastructure/image_command_repository.go`
- Add: `code/backend/internal/module/challenge/infrastructure/image_command_repository_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/image_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/image_service_test.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/image_service_test.go`

- [x] 为 `CreateImage` 增加模块 sentinel not-found 仍允许创建的测试
- [x] 为 `UpdateImage` / `DeleteImage` 增加模块 sentinel -> `errcode.ErrImageNotFound` 的测试
- [x] 跑最小测试，确认红灯来自 image service 尚未消费模块 sentinel

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'ImageService' -count=1 -timeout 300s`

## Task 2: 实现 adapter 与 wiring

**Files:**
- Add: `code/backend/internal/module/challenge/infrastructure/image_command_repository.go`
- Add: `code/backend/internal/module/challenge/infrastructure/image_command_repository_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/image_service.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`

- [x] 新增 command-side image adapter，把 raw repo 的 `gorm.ErrRecordNotFound` 映射成 `ErrChallengeImageNotFound`
- [x] 让 `ImageService` 改成只看模块 sentinel
- [x] 在 runtime wiring 中只给 image command service 注入 adapter
- [x] 为 adapter 增加 raw GORM not-found -> sentinel 与 passthrough 测试

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'ImageService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'Image(Command|Query)Repository' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s`

## Risks

- `CreateImage` 的 not-found 语义是“继续创建”，不能被误映射成 `ErrImageNotFound`
- image build 仍依赖 raw image repo，runtime 注入面必须只收口 image command 路径
- 复用 `ErrChallengeImageNotFound` 时不能顺手改变 image query 已有语义

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'ImageService' -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'Image(Command|Query)Repository' -count=1 -timeout 300s`
3. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s`
