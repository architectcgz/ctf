# Challenge Query Not-Found Contract Phase 5 Slice 31 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `challenge/application/queries/challenge_service.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持 challenge lookup 在 `GetChallenge` 与 `GetPublishedChallenge` 上的既有不同 `errcode` 语义不变。

**Architecture:** 新增一层窄 `ChallengeQueryRepository` adapter，把 raw challenge query repository 的 `gorm.ErrRecordNotFound` 收口成 `challenge/ports` sentinel；`ChallengeService` 只消费 sentinel，再按入口分别映射成 `errcode.ErrChallengeNotFound` 与 `errcode.ErrNotFound`。runtime builder 负责注入 adapter。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `challenge/application/queries/challenge_service.go -> gorm.io/gorm`
- 保持 `GetChallenge` 的 challenge lookup not-found -> `errcode.ErrChallengeNotFound`
- 保持 `GetPublishedChallenge` 的同一 lookup not-found -> `errcode.ErrNotFound`

## Non-goals

- 不修改 `code/backend/internal/module/architecture_allowlist_test.go`
- 不修改 `docs/architecture/backend/07-modular-monolith-refactor.md`
- 不修改 `docs/design/backend-module-boundary-target.md`
- 不修改 `.harness/reuse-decisions/echarts-mount-gate.md`
- 不处理 challenge command / topology / image 其他 concrete 依赖
- 不改 contest 模块代码

## Inputs

- `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Ownership Boundary

- `challenge/application/queries/challenge_service.go`
  - 负责：challenge query 编排、fallback 和 application `errcode` 映射
  - 不负责：知道底层 not-found 是否来自 GORM
- `challenge/ports/ports.go`
  - 负责：定义 challenge query lookup 的模块内 sentinel
  - 不负责：决定 HTTP / API 对外错误码
- `challenge/infrastructure/challenge_query_repository.go`
  - 负责：把 raw repository 的 challenge lookup not-found 映射成 sentinel
  - 不负责：改变 list / stats / cache / hints 语义
- `challenge/runtime/module.go`
  - 负责：把 adapter 注入 query service
  - 不负责：把 raw GORM concrete 带回 application surface

## Change Surface

- Add: `.harness/reuse-decisions/challenge-query-not-found-contract-phase5-slice31.md`
- Add: `docs/plan/impl-plan/2026-05-13-challenge-query-not-found-contract-phase5-slice31-implementation-plan.md`
- Add: `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
- Add: `code/backend/internal/module/challenge/infrastructure/challenge_query_repository_test.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- Add: `code/backend/internal/module/challenge/infrastructure/challenge_query_repository_test.go`

- [ ] 为 `GetChallenge` / `GetPublishedChallenge` 增加 sentinel 分支测试，锁定两个入口必须继续映射成不同 `errcode`
- [ ] 为 `ChallengeQueryRepository` 增加 raw GORM not-found -> ports sentinel 的映射测试
- [ ] 跑最小测试，确认红灯来自目标 contract 尚未实现

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/queries -run 'ChallengeService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'Challenge.*Repository|ChallengeQuery' -count=1 -timeout 300s`

## Task 2: 实现 adapter 与 wiring

**Files:**
- Add: `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`

- [ ] 在 `challenge/ports` 增加 query lookup not-found sentinel
- [ ] 新增 query adapter，把 raw repo 的 `gorm.ErrRecordNotFound` 映射成 sentinel
- [ ] 让 `ChallengeService` 改成只看 sentinel，并保持两个入口的不同 `errcode`
- [ ] 在 runtime wiring 中给 query service 注入 adapter

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/queries -run 'ChallengeService' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'Challenge.*Repository|ChallengeQuery' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s`

## Risks

- 同一个 challenge lookup not-found 在 `GetChallenge` / `GetPublishedChallenge` 上有不同 `errcode`，不能因为 sentinel 收口而把两个入口压平
- `ChallengeService` 同时承担 cache fallback、stats fallback 和 hints 读取，adapter 只能承接 lookup not-found，不能顺手改其他失败边界
- `ports.go` 与 `runtime/module.go` 当前工作区已有其他 slice 的未提交修改，本次只能在已读上下文上做最小增量

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/queries -run 'ChallengeService' -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -run 'Challenge.*Repository|ChallengeQuery' -count=1 -timeout 300s`
3. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s`
