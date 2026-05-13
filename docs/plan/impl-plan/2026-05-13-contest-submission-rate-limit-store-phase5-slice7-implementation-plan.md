# Contest Submission Rate Limit Store Phase 5 Slice 7 Implementation Plan

## Objective

继续 phase 5 收窄 `contest` application concrete allowlist，把 `submission` 错误提交限流的 Redis 依赖下沉到模块 port / infrastructure adapter：

- 去掉 `contest/application/commands/submission_service.go -> github.com/redis/go-redis/v9`
- 保持错误提交限流 key、TTL、失败时中断提交、正确提交不写限流的现有行为不变

## Non-goals

- 不处理 `AWDService`、`contest_awd_service_service`、`awd_checker_preview_token_support.go` 等更宽的 AWD Redis 状态面
- 不改 scoreboard 增量更新、提交流水线事务或 flag 校验语义
- 不改错误提交限流 key 形状、默认 prefix 或配置字段

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/contest/application/commands/submission_*.go`
- `code/backend/internal/module/contest/ports/submission.go`
- `code/backend/internal/module/contest/runtime/module.go`

## Current Baseline

- `SubmissionService` 当前直接持有 `*redis.Client`
- `validateContestSubmission` 直接 `Exists` 查询错误提交限流 key
- `handleIncorrectSubmission` 直接 `Set` 写入错误提交限流 key 与 TTL
- allowlist 当前保留 `contest/application/commands/submission_service.go -> github.com/redis/go-redis/v9`

## Chosen Direction

把错误提交限流状态面表达成 `contest` 自己的 submission rate-limit store：

1. 在 `contest/ports` 新增 `ContestSubmissionRateLimitStore`
2. `SubmissionService` 只依赖该 store，保留“什么时候查限流、什么时候写限流、失败时中断提交”的业务编排语义
3. 在 `contest/infrastructure` 新增 Redis adapter，内部封装 rate-limit key 与 `Exists/Set` 细节
4. `runtime/module.go` 统一构建该 store，并注入 submission service

## Ownership Boundary

- `contest/application/commands/submission_*.go`
  - 负责：校验提交流程、决定何时查错误提交限流、何时写入限流、何时继续评分或中断
  - 不负责：知道 Redis client、`Exists`、`Set` 或 key 默认前缀回退细节
- `contest/infrastructure/submission_rate_limit_store.go`
  - 负责：实现 Redis 限流状态存取、构造 key、处理默认 prefix
  - 不负责：决定提交流程或何时触发限流
- `contest/runtime/module.go`
  - 负责：装配 configured rate-limit store 并注入 submission service
  - 不负责：把 Redis client 继续暴露回 submission application surface

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-13-contest-submission-rate-limit-store-phase5-slice7-implementation-plan.md`
- Add: `.harness/reuse-decisions/contest-submission-rate-limit-store-phase5-slice7.md`
- Add: `code/backend/internal/module/contest/infrastructure/submission_rate_limit_store.go`
- Add: `code/backend/internal/module/contest/ports/submission_rate_limit_context_contract_test.go`
- Modify: `code/backend/internal/module/contest/ports/submission.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_submit.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_submit_validation.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_incorrect_submit.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

- [x] Slice 1: 提取 submission rate-limit store port 与 Redis adapter
  - 目标：`SubmissionService` 不再持有 `*redis.Client`，错误提交限流仍保持现有 key / TTL / failure behavior
  - 验证：
    - `cd code/backend && go test ./internal/module/contest/application/commands -run 'SubmissionService' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`
  - Review focus：submission application 是否已经不再知道 Redis concrete；配置 prefix 与默认 prefix 是否保持一致

- [x] Slice 2: 删除 allowlist 并同步文档
  - 目标：删除本次 submission 实际收口的 allowlist，并更新 phase 5 当前事实
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除 `submission_service.go` 这条例外；文档准确描述 submission rate-limit 已下沉，但 AWD Redis 依赖仍未完成

## Risks

- 如果 key 构造与默认 prefix 回退不一致，现有 rate-limit 行为会回归
- 如果错误提交写限流失败后没有继续中断提交流程，会改变当前失败语义
- 如果 runtime 继续把旧 Redis client 注回 submission service，allowlist 不会真正收口

## Verification Plan

1. `cd code/backend && go test ./internal/module/contest/application/commands -run 'SubmissionService' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：submission application 继续持有提交流程与限流编排语义，Redis 限流 key / storage 细节落回 infrastructure
- reuse point 明确：沿用 phase5 既有 port + state-store 模式，不引入新的宽 generic cache helper
- 这刀同时解决行为与结构：保留当前限流行为，同时删掉 `submission_service` 的 concrete Redis import
