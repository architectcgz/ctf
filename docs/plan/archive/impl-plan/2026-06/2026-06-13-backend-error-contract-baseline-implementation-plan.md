<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# Backend error contract baseline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking; flip each checkbox immediately after the expected result is reached.

**Goal:** 建立后端 public app error 与 transport/infrastructure sentinel 的边界基线，并迁移一个真实的 GORM not-found 泄漏点。

**Architecture:** 保持现有 modular-monolith / Onion 方向：infrastructure adapter 负责把 concrete persistence sentinel 映射为模块可消费语义或 public contract error，HTTP transport 只通过 `httpresponse.FromError` 消费 public `apperror.AppError`。本 slice 只建立 guardrail 和 challenge owner guard 试点，不新增泛化 database/cache/docker app error，也不全量迁移 `ErrInternal`。

**Tech Stack:** Go, Gin, GORM, `apperror`, source architecture tests, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-13-backend-error-contract-baseline`
- Parent Task Group: `2026-06-13-backend-error-management-group`
- Slice Index: `2/13`
- Depends On: `无`
- Started At: `2026-06-13T00:13:47Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-error-contract-baseline`
- Branch: `task/2026-06-13-backend-error-contract-baseline`
- Plan Type: `slice`

## Plan Status

- Status: `review-passed`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective:
  - 给 public `apperror.AppError` 增加基础回归测试，确认 `WithCause` / `WithMessage` 不改变 public code、message、HTTP status 和 `errors.Is` 语义。
  - 让 `challenge/infrastructure.ContractRepository.FindByID` 把 raw `gorm.ErrRecordNotFound` 映射为 `challenge/contracts.ErrChallengeNotFound`。
  - 移除 `internal/app/router_routes.go` 中对 `gorm.ErrRecordNotFound` 的直接分支，让 `challengeOwnerGuard` 只调用 `response.FromError`。
  - 增加源码级 guardrail，禁止非装配层 transport/API/middleware 直接消费 persistence/cache/runtime sentinel。
- Non-Goals:
  - 不实现父计划中 `ErrDatabaseConnectionFailed` / `ErrCacheTimeout` / `ErrDockerAPIFailed` 等泛化错误码；这些需要先有 owner 和调用语义。
  - 不全仓库迁移 `ErrInternal`。
  - 不改变已有 API JSON envelope 格式。
  - 不移动 app composition root 中合法的 `*gorm.DB` / Redis / Docker runtime wiring。
  - 不修改数据库 schema、migration、Redis key 或 Docker runtime 行为。

## Problem Statement

- Current behavior / structure:
  - `internal/httpresponse.FromError` 已经能把 `*apperror.AppError` 映射成 public response，并把 non-app error 统一映射为 `ErrInternal`。
  - `internal/app/router_routes.go` 的 `challengeOwnerGuard` 仍 import `gorm.io/gorm` 并分支 `gorm.ErrRecordNotFound`，HTTP transport 知道了 persistence sentinel。
  - `challenge/infrastructure.ContractRepository.FindByID` 直接透传 raw repository error，导致 app 层只能识别 GORM sentinel。
  - 现有 architecture test 还没有禁止 app/api/middleware transport 继续新增这类分支。
- Target behavior / structure:
  - challenge contract repository 把 GORM not-found 映射为 public challenge contract error。
  - router guard 对 lookup error 统一走 `response.FromError`，不再知道 GORM。
  - architecture guardrail 能在源码层拦截 transport/API/middleware 的 sentinel import / branch 回流。
- Why this task is needed now:
  - 后续 Redis/Docker/error migration slices 都依赖一个可审查的错误边界基线；先用一个小试点建立模式和 guardrail，避免后续继续扩大 transport 层泄漏。

## Inputs

- Source docs:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-backend-error-management-improvement-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/tests/README.md`
  - `code/backend/internal/httpresponse/response.go`
  - `code/backend/internal/apperror/error.go`
- Related prior work:
  - `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
  - `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
  - `code/backend/internal/module/container_runtime/infrastructure/engine_errors.go`
  - Slice 1 commit `0c745cb84` for architecture guardrail style.

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 触达 HTTP transport、module contract repository、public app error contract 和源码级 architecture guardrail。
  - 需要 TDD 证明 not-found 映射和 boundary guardrail 生效，并需要独立 review gate。

## Files

- Create:
  - `code/backend/internal/module/challenge/infrastructure/contract_repository_test.go`
- Modify:
  - `code/backend/internal/apperror/error_test.go`
  - `code/backend/internal/module/challenge/infrastructure/contract_repository.go`
  - `code/backend/internal/module/contest/infrastructure/contest_challenge_lookup_adapter.go`
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/tests/architecture/test_architecture_test.go`
  - `docs/architecture/backend/04-api-design.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-contract-baseline-implementation-plan.md`
- Review:
  - `code/backend/internal/httpresponse/response.go`
  - `code/backend/internal/module/challenge/contracts/errors.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
  - `code/backend/internal/app/router.go`
  - `code/backend/internal/app/composition/*.go`
- Test:
  - `code/backend/internal/apperror/error_test.go`
  - `code/backend/internal/module/challenge/infrastructure/contract_repository_test.go`
  - `code/backend/internal/module/contest/infrastructure/contest_challenge_lookup_adapter_test.go`
  - `code/backend/tests/architecture/test_architecture_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `challenge/infrastructure/challenge_query_repository.go` 已将 raw `gorm.ErrRecordNotFound` 映射为 `challenge/ports.ErrChallengeQueryChallengeNotFound`。
  - `challenge/infrastructure/flag_repository.go`、`topology_service_repository.go`、`contest/infrastructure/*` 多处已有 adapter 层 sentinel 映射。
  - `httpresponse.FromError` 已是 public app error 到 HTTP envelope 的统一响应边界。
  - `tests/architecture/test_architecture_test.go` 已承担源码级 guardrail。
- Reuse / extend / split / create-new decision:
  - 复用 `challengecontracts.ErrChallengeNotFound`，不新增同义错误码。
  - 直接修改 `ContractRepository`，因为它就是 challenge 对外 contract adapter，`FindByID` 的调用方都应收到 public contract error。
  - 在 `tests/architecture` 增加 transport sentinel guardrail，不在 `internal/app` 自测里新增只读源码扫描。
- Owner boundary:
  - `challenge/infrastructure.ContractRepository`：raw challenge repository -> challenge public contract error 映射 owner。
  - `internal/app/router_routes.go`：路由权限 guard，只消费 public error，不识别 GORM/Redis/Docker sentinel。
  - `httpresponse.FromError`：HTTP envelope 映射 owner。
  - `tests/architecture`：长期防回流 guardrail owner。
- Why this is the narrowest safe surface:
  - 只迁移一个已确认泄漏点和一个 contract adapter，不改变 repository 全局语义。
  - Guardrail 精确扫描 transport/API/middleware runtime 源码，跳过 `_test.go` 和 `internal/app/composition` 装配路径，避免误伤合法 wiring。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 原父计划范围很大，必须先拆成可独立 review 的最小错误边界试点。
- grill-with-docs findings:
  - 架构事实已明确 `application` / `api` / transport 不应直接知道 `gorm.ErrRecordNotFound`，现有 `challengeOwnerGuard` 与该事实冲突。
  - `internal/app/router.go`、`http_server.go` 和 `internal/app/composition` 合法承担 process-level wiring，不能简单禁止整个 app package import GORM。
  - 父计划中的泛化 infra app error 还没有明确 owner 和调用语义；本 slice 先不落地。
- Plan adjustments after challenge:
  - 将目标收窄到 app error contract baseline、challenge contract adapter 试点和 source guardrail。
  - Guardrail 关注 concrete sentinel 消费，不禁止合法装配层导入 concrete clients。
  - 新增 AppError 基线测试，但不改 public error registry 运行时设计。

## Execution Slices

### Slice 1: Public AppError baseline tests

- Goal: 固化现有 AppError public contract 行为，避免后续错误细分时破坏 `errors.Is` / response status。
- Dependencies: 无
- Files:
  - Create:
  - Modify:
    - `code/backend/internal/apperror/error_test.go`
  - Review:
    - `code/backend/internal/apperror/error.go`
  - Test:
    - `code/backend/internal/apperror/error_test.go`
- Steps:
  - [x] Step 1: 写 `TestAppErrorWithCausePreservesPublicContract` 和 `TestAppErrorWithMessagePreservesCodeAndHTTPStatus`。
  - [x] Step 2: 运行 `cd code/backend && go test ./internal/apperror -run 'TestAppErrorWith(CausePreservesPublicContract|MessagePreservesCodeAndHTTPStatus)' -count=1`，确认测试覆盖当前 contract。
  - [x] Step 3: 如测试暴露 contract 缺口，只做最小修复；若现有实现已满足，保持生产代码不动。
- Validation: `cd code/backend && go test ./internal/apperror -count=1`
- Review focus: 不让 cause 或 message override 改变 code/status；不把 cause 暴露成 public message。
- Done criteria: AppError focused tests 通过。

### Slice 2: Challenge contract repository maps not-found

- Goal: 用 TDD 把 raw GORM not-found 映射到 challenge public contract error。
- Dependencies: Slice 1
- Files:
  - Create:
    - `code/backend/internal/module/challenge/infrastructure/contract_repository_test.go`
  - Modify:
    - `code/backend/internal/module/challenge/infrastructure/contract_repository.go`
  - Review:
    - `code/backend/internal/module/challenge/contracts/errors.go`
    - `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
  - Test:
    - `code/backend/internal/module/challenge/infrastructure/contract_repository_test.go`
- Steps:
  - [x] Step 4: 写 `TestContractRepositoryMapsFindByIDNotFoundToChallengeContractError`，先期望 `errors.Is(err, challengecontracts.ErrChallengeNotFound)`。
  - [x] Step 5: 运行 `cd code/backend && go test ./internal/module/challenge/infrastructure -run TestContractRepositoryMapsFindByIDNotFoundToChallengeContractError -count=1`，确认失败。
  - [x] Step 6: 在 `ContractRepository.FindByID` 中用 `errors.Is(err, gorm.ErrRecordNotFound)` 映射 `challengecontracts.ErrChallengeNotFound`。
  - [x] Step 7: 增加或运行 passthrough 测试，确认非 not-found error 不被吞掉。
  - [x] Step 8: 重跑 `cd code/backend && go test ./internal/module/challenge/infrastructure -run TestContractRepository -count=1`。
- Validation: `cd code/backend && go test ./internal/module/challenge/infrastructure -run TestContractRepository -count=1`
- Review focus: 只改 contract adapter 语义，不影响 raw `Repository.FindByID` 的其他调用方。
- Done criteria: contract repository not-found / passthrough tests 通过。

### Slice 3: Transport sentinel guardrail and router cleanup

- Goal: 移除 router owner guard 的 GORM sentinel 分支，并添加源码 guardrail 防止回流。
- Dependencies: Slice 2
- Files:
  - Create:
  - Modify:
    - `code/backend/internal/app/router_routes.go`
    - `code/backend/tests/architecture/test_architecture_test.go`
  - Review:
    - `code/backend/internal/httpresponse/response.go`
    - `code/backend/internal/app/router.go`
    - `code/backend/internal/app/composition/*.go`
  - Test:
    - `code/backend/tests/architecture/test_architecture_test.go`
- Steps:
  - [x] Step 9: 写 `TestTransportLayersDoNotConsumePersistenceOrRuntimeSentinels`，扫描非测试 Go 文件中的 `internal/app` 非 composition/runtime、`internal/module/*/api` 和 `internal/middleware`。
  - [x] Step 10: 运行 `cd code/backend && go test ./tests/architecture -run TestTransportLayersDoNotConsumePersistenceOrRuntimeSentinels -count=1`，确认当前 `router_routes.go` 失败。
  - [x] Step 11: 从 `router_routes.go` 删除 `errors` / `gorm.io/gorm` import 和 special-case 分支，lookup error 统一 `response.FromError(c, err)`。
  - [x] Step 12: 重跑 architecture guardrail，确认通过。
  - [x] Step 13: 运行 `cd code/backend && go test ./internal/app -run 'Test.*ChallengeOwner|TestArchitectureRules' -count=1`，确认 app package 编译与相关测试通过；若没有匹配测试，记录 no tests to run。
- Validation:
  - `cd code/backend && go test ./tests/architecture -run TestTransportLayersDoNotConsumePersistenceOrRuntimeSentinels -count=1`
  - `cd code/backend && go test ./internal/app -run 'Test.*ChallengeOwner|TestArchitectureRules' -count=1`
- Review focus: guardrail 不误伤 app composition/root wiring；handler/API/middleware 不再分支 GORM/Redis/Docker sentinel。
- Done criteria: router cleanup 编译通过，architecture guardrail 通过。

### Slice 4: Workflow validation, review evidence, and docs state

- Goal: 跑当前 slice 的最小充分验证，更新 task group index 和 review handoff。
- Dependencies: Slice 1-3
- Files:
  - Create:
    - `docs/reviews/backend/2026-06-13-backend-review-error-contract-baseline.md`
  - Modify:
    - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
    - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-contract-baseline-implementation-plan.md`
  - Review:
  - Test:
- Steps:
  - [x] Step 14: 运行 `cd code/backend && go test ./internal/apperror ./internal/module/challenge/infrastructure ./tests/architecture -run 'Test(AppError|ContractRepository|TransportLayers)' -count=1`。
  - [x] Step 15: 运行 `git diff --check -- <touched files>`。
  - [x] Step 16: 运行 `bash scripts/check-startup-gate.sh`。
  - [x] Step 17: 运行 `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`。
  - [x] Step 18: 运行 `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`。
  - [x] Step 19: 提交本 slice，commit message 包含正文两行和 `Task: 2026-06-13-backend-error-contract-baseline`。
  - [x] Step 20: 运行独立 `code-reviewer` gate，归档 review 到 `docs/reviews/backend/2026-06-13-backend-review-error-contract-baseline.md`。
  - [x] Step 21: 处理 material findings 后重跑受影响验证；无 findings 时更新 plan、group index 和 review link。
  - [x] Step 22: 运行 `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`。
- Validation: 见上。
- Review focus: error boundary owner、guardrail 精度、public error response 兼容。
- Done criteria: focused tests、workflow stages 和 independent review gate 通过；plan/group index/review evidence 同步。

## Impact And Compatibility

- API / DTO:
  - JSON envelope 不变。
  - Challenge not-found response 仍为 `challengecontracts.ErrChallengeNotFound` 对应的 404 / code `13004`。
- Data / migration:
  - 无 schema 或 migration 变更。
- State / cache / queue / event:
  - 无 Redis/cache/queue/event 行为变更。
- Runtime / config:
  - 无配置或运行时依赖变更。
- Frontend route / state / UX:
  - 无前端改动；用户可见错误码和 message 保持不变。
- Docs / contracts:
  - 更新 `docs/architecture/backend/04-api-design.md`，记录本 slice 不改变 HTTP path/schema/envelope，只把 challenge not-found 的内部错误边界收回 module contract adapter。
  - 更新 task group index 和当前 slice plan；不把本 slice 计划写成架构事实源。

## Plan Review / Architecture Fit

- Target owner boundary:
  - infrastructure adapter 消费 concrete persistence sentinel；transport 消费 public app error；response helper 负责 HTTP envelope。
- Reuse points / landing zones:
  - `httpresponse.FromError` 是 existing public error response boundary。
  - `tests/architecture` 是 source guardrail 落点。
  - `challengecontracts.ErrChallengeNotFound` 是 existing public not-found contract。
- Known structural debt touched:
  - 触达 `internal/app/router_routes.go` 中已知 transport -> GORM sentinel 泄漏。
- How this plan avoids behavior-only convergence:
  - 不只删除一处分支，还新增长期 guardrail 防止相同泄漏回流。
- Hidden second-redesign risk:
  - 中低。后续 Redis/Docker slices 会复用 guardrail 和 adapter 模式；如果需要更细的 infra app error，应在对应 slice 结合实际 owner 新增。
- Decision after review:
  - `ready-for-implementation`。

## Documentation Owner

- Current fact sources to read:
  - `docs/文档规范.md`
  - `docs/plan/README.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
- Fact sources to update after implementation:
  - 本 slice 更新 implementation plan、task group index 和 review evidence。
  - 若 review 发现架构事实源缺少当前已落地的 guardrail，可在后续 docs/runbook slice 统一吸收。
- Plan-only notes that must not become architecture source:
  - 父计划中的泛化 infra app error code 只是候选方向，不是当前落地事实。
- Archive condition:
  - 当前 slice 合并并通过 independent review 后归档当前 slice plan；task group index 保持活动直到 task group 完成。

## Validation Plan

- Per-slice commands:
  - `cd code/backend && go test ./internal/apperror -count=1`
  - `cd code/backend && go test ./internal/module/challenge/infrastructure -run TestContractRepository -count=1`
  - `cd code/backend && go test ./tests/architecture -run TestTransportLayersDoNotConsumePersistenceOrRuntimeSentinels -count=1`
  - `cd code/backend && go test ./internal/app -run 'Test.*ChallengeOwner|TestArchitectureRules' -count=1`
- Integration commands:
  - `cd code/backend && go test ./internal/apperror ./internal/module/challenge/infrastructure ./tests/architecture -run 'Test(AppError|ContractRepository|TransportLayers)' -count=1`
  - `bash scripts/check-startup-gate.sh`
  - `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`
  - `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`
- Manual checks:
  - Diff 中 `router_routes.go` 不再 import `gorm.io/gorm` 或 `errors`。
  - Guardrail allowlist 不包含当前已修复文件。
- Commands intentionally skipped and why:
  - 不跑全量 backend package tests；当前 slice 只触达 apperror、challenge infrastructure、router compile 和 source architecture guardrail，workflow completion-full 已覆盖项目关键机械检查。

## Validation Evidence

- Command: `cd code/backend && go test ./internal/apperror -run 'TestAppErrorWith(CausePreservesPublicContract|MessagePreservesCodeAndHTTPStatus)' -count=1`
  - Result: PASS
  - Notes: Existing AppError implementation already preserved public code/message/status and cause inspection.
- Command: `cd code/backend && go test ./internal/module/challenge/infrastructure -run TestContractRepositoryMapsFindByIDNotFoundToChallengeContractError -count=1`
  - Result: FAIL
  - Notes: Red step failed with raw `record not found`, proving the contract repository still leaked GORM not-found.
- Command: `cd code/backend && go test ./internal/module/challenge/infrastructure -run TestContractRepository -count=1`
  - Result: PASS
  - Notes: Green step maps GORM not-found to `challengecontracts.ErrChallengeNotFound` and passes through non-not-found infrastructure errors.
- Command: `cd code/backend && go test ./tests/architecture -run TestTransportLayersDoNotConsumePersistenceOrRuntimeSentinels -count=1`
  - Result: FAIL
  - Notes: Red step failed on `internal/app/router_routes.go:72` consuming `gorm.ErrRecordNotFound`.
- Command: `cd code/backend && go test ./tests/architecture -run TestTransportLayersDoNotConsumePersistenceOrRuntimeSentinels -count=1`
  - Result: PASS
  - Notes: Green step passed after `challengeOwnerGuard` stopped branching on GORM sentinel.
- Command: `cd code/backend && go test ./internal/app -run 'Test.*ChallengeOwner|TestArchitectureRules' -count=1`
  - Result: PASS
  - Notes: App package compiles with the simplified guard.
- Command: `cd code/backend && go test ./internal/apperror ./internal/module/challenge/infrastructure ./tests/architecture -run 'Test(AppError|ContractRepository|TransportLayers)' -count=1`
  - Result: PASS
  - Notes: Combined focused validation for this slice.
- Command: `git diff --check -- code/backend/internal/apperror/error_test.go code/backend/internal/module/challenge/infrastructure/contract_repository.go code/backend/internal/module/challenge/infrastructure/contract_repository_test.go code/backend/internal/app/router_routes.go code/backend/tests/architecture/test_architecture_test.go docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-contract-baseline-implementation-plan.md`
  - Result: PASS
  - Notes: No whitespace errors.
- Command: `bash scripts/check-startup-gate.sh`
  - Result: PASS
  - Notes: Startup gate is active for `2026-06-13-backend-error-contract-baseline`.
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`
  - Result: PASS
  - Notes: Frontend test guard, startup gate, backend/frontend architecture, and backend test architecture guard passed.
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: FAIL
  - Notes: `10-code-change-contracts.sh` flagged `internal/app/router_routes.go` as API surface changed without contract documentation. Architecture plugins passed. Root cause was the project guard treating any router surface change conservatively; `docs/architecture/backend/04-api-design.md` was updated to record that HTTP contract stays unchanged and only internal error boundary changed.
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: PASS
  - Notes: Passed after updating `docs/architecture/backend/04-api-design.md`; code change contracts, backend architecture, frontend architecture and test architecture guard passed.
- Command: `bash scripts/check-architecture.sh --full`
  - Result: PASS
  - Notes: Full backend/frontend architecture checks passed after the API design doc update.
- Command: `timeout 600s codex exec --sandbox read-only ...`
  - Result: BLOCKED
  - Notes: Independent review found that `ContestChallengeLookupAdapter` did not map the new upstream `challengecontracts.ErrChallengeNotFound` shape back to `contestports.ErrContestChallengeEntityNotFound`, which could turn missing challenge paths in contest commands into public 500 responses.
- Command: `cd code/backend && go test ./internal/module/contest/infrastructure -run TestContestChallengeLookupAdapterMapsNotFoundErrors -count=1`
  - Result: FAIL
  - Notes: Red step for the review finding failed on the new `public challenge contract error` case.
- Command: `cd code/backend && go test ./internal/module/contest/infrastructure -run TestContestChallengeLookupAdapter -count=1`
  - Result: PASS
  - Notes: Adapter now maps `challengecontracts.ErrChallengeNotFound` to `contestports.ErrContestChallengeEntityNotFound`.
- Command: `cd code/backend && go test ./internal/module/contest/application/commands -run 'TestChallengeServiceAddChallengeToContestTreatsChallengeSentinelAsErrChallengeNotFound|TestContestAWDServiceSyncContestChallengeRelationTreatsChallengeSentinelAsErrChallengeNotFound' -count=1`
  - Result: PASS
  - Notes: Contest command public not-found behavior remains `challengecontracts.ErrChallengeNotFound`.
- Command: `cd code/backend && go test ./internal/apperror ./internal/module/challenge/infrastructure ./tests/architecture -run 'Test(AppError|ContractRepository|TransportLayers)' -count=1`
  - Result: PASS
  - Notes: Original slice focused validation still passes after the contest adapter compatibility fix.
- Command: `git diff --check -- <touched files>`
  - Result: PASS
  - Notes: Re-run after the contest adapter compatibility fix.
- Command: `bash scripts/check-startup-gate.sh`
  - Result: PASS
  - Notes: Re-run after the contest adapter compatibility fix.
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: PASS
  - Notes: Re-run after the contest adapter compatibility fix; code changes, backend architecture, frontend architecture no-op, and test architecture guard passed.
- Command: `timeout 420s codex exec --sandbox read-only --cd /home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-error-contract-baseline --output-last-message /tmp/ctf-error-contract-baseline-review.md`
  - Result: PASS
  - Notes: Independent `code-reviewer` gate returned `pass` with no material findings. Review archived at `docs/reviews/backend/2026-06-13-backend-review-error-contract-baseline.md`.
- Command: `cd code/backend && go test ./internal/apperror ./internal/module/challenge/infrastructure ./internal/module/contest/infrastructure ./internal/module/contest/application/commands ./tests/architecture -run 'Test(AppError|ContractRepository|TransportLayers|ContestChallengeLookupAdapter|ChallengeServiceAddChallengeToContestTreatsChallengeSentinelAsErrChallengeNotFound|TestContestAWDServiceSyncContestChallengeRelationTreatsChallengeSentinelAsErrChallengeNotFound)' -count=1`
  - Result: PASS
  - Notes: Fresh focused verification on the branch after merging latest `main`.
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: PASS
  - Notes: Fresh completion gate on the branch after merging latest `main`; API contract surface unchanged.
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`
  - Result: PASS
  - Notes: Governance audit passed after archiving review evidence and syncing the branch with latest `main`.

## Independent Review Handoff

- Review target:
  - Commit `c00f89954` initially reviewed as `HEAD~1..HEAD`; review returned blocked.
  - Final review target after fix: current `HEAD~1..HEAD`.
- Validation evidence summary:
  - Focused AppError, challenge contract repository, transport architecture, contest adapter and contest command tests passed.
  - `pre-commit-quick`, `completion-full` and `check-architecture.sh --full` passed before first review; `completion-full` must be rerun after the contest adapter fix.
- Architecture / contract inputs:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/tests/README.md`
  - `code/backend/internal/httpresponse/response.go`
  - `code/backend/internal/apperror/error.go`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
- Known risks / review focus:
  - Guardrail 是否误伤 app composition root 的 legitimate wiring。
  - `ContractRepository` 映射 public error 是否改变其他跨模块调用期望。
  - HTTP not-found response 是否仍保持 404 / code `13004`。
- Project-local checks to consider:
  - `bash scripts/check-startup-gate.sh`
  - `cd code/backend && go test ./internal/apperror ./internal/module/challenge/infrastructure ./tests/architecture -run 'Test(AppError|ContractRepository|TransportLayers)' -count=1`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`

## Review Gate Status

- Independent review gate: `pass`
- Review archive: `docs/reviews/backend/2026-06-13-backend-review-error-contract-baseline.md`
- Notes:
  - First independent review blocked on `ContestChallengeLookupAdapter` not mapping the new public challenge sentinel back to contest port semantics.
  - The implementation context fixed that compatibility gap, re-ran the focused contest adapter / command tests and `completion-full`, then requested a fresh independent re-review.
  - Final independent read-only review returned `pass` with no material findings.
- Required next action: 运行 `workflow-governance`，随后归档当前 slice plan 并把 startup gate 推进到 `ready_to_merge`。
- Required next action: 归档当前 slice plan 并把 startup gate 推进到 `ready_to_merge`。

## Rollback / Recovery

- Safe revert boundary:
  - 单个 task commit 可安全 revert；无数据或配置变更。
- Data / config / runtime recovery notes:
  - 无。
- Irreversible operations:
  - 无。

## Residual Risks

- Risk:
  - Guardrail 只覆盖 transport/API/middleware 的 concrete sentinel 消费，不覆盖所有 application service。
- Why acceptable:
  - 本 slice 目标是 HTTP transport boundary baseline；后续 Redis/Docker/application migration slices 会逐步扩大覆盖面。
- Follow-up owner, if any:
  - `backend-redis-error-boundary`
  - `backend-container-runtime-error-boundary`
  - `backend-application-error-migration-core`
