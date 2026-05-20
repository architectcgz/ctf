# Contest AWD Runtime State And Preview Token Store Phase 5 Slice 8 Implementation Plan

## Objective

继续 phase 5 收窄 `contest` application concrete allowlist，把 AWD 剩余 Redis 状态面下沉到模块 port / infrastructure adapter：

- 去掉 `contest/application/commands/awd_checker_preview_token_support.go -> github.com/redis/go-redis/v9`
- 去掉 `contest/application/commands/awd_current_round_fallback_support.go -> github.com/redis/go-redis/v9`
- 去掉 `contest/application/commands/awd_flag_support.go -> github.com/redis/go-redis/v9`
- 去掉 `contest/application/commands/awd_service.go -> github.com/redis/go-redis/v9`
- 去掉 `contest/application/commands/awd_status_cache.go -> github.com/redis/go-redis/v9`
- 去掉 `contest/application/commands/contest_awd_service_service.go -> github.com/redis/go-redis/v9`
- 去掉 `contest/application/commands/contest_awd_service_support.go -> github.com/redis/go-redis/v9`

保持 AWD 当前轮次 fallback、flag 解析、service status cache、checker preview token 存取与消费的现有行为不变。

## Non-goals

- 不改 AWD 业务规则、checker 执行流程、评分 SQL 或 scoreboard cache 行为
- 不改 `practice`、`challenge` 等其他模块
- 不重写 AWD checker preview token 的匹配语义

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/contest/application/commands/awd_*.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_*.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/runtime/module.go`

## Current Baseline

- `AWDService` 当前直接持有 `*redis.Client`
- current round fallback、round flag 读取、service status 写入都散落在 AWD application 里直接操作 Redis
- checker preview token 的 `Set/Get/Del` 与 JSON 编解码散落在 application helper，并被 `AWDService` 与 `ContestAWDServiceService` 共用
- allowlist 当前还保留上述 7 条 contest AWD Redis 例外

## Chosen Direction

把 AWD 剩余 Redis 状态面拆成两个窄 store：

1. 扩展现有 `contest/ports.AWDRoundStateStore`，承接 current round number 读取与单字段 service status 写入
2. 新增 `contest/ports.AWDCheckerPreviewTokenStore`，承接 checker preview token 的持久化、加载与删除
3. `AWDService` 只依赖 `AWDRoundStateStore + AWDCheckerPreviewTokenStore`，不再持有 `*redis.Client`
4. `ContestAWDServiceService` 只依赖 `AWDCheckerPreviewTokenStore`
5. `contest/infrastructure` 提供 Redis adapter；`runtime/module.go` 统一装配并注入 AWD command surfaces

## Ownership Boundary

- `contest/application/commands/awd_*.go`
  - 负责：决定何时 fallback 当前轮次、何时读 flag、何时更新 service status、何时保存 preview token
  - 不负责：知道 Redis key、`redis.Nil`、`HGet`、`Get`、`Set` 或 `Del`
- `contest/application/commands/contest_awd_service_*.go`
  - 负责：基于 preview token 决定 validation state 是否更新
  - 不负责：知道 token 在 Redis 里的 key、JSON 序列化或删除细节
- `contest/infrastructure`
  - 负责：实现 AWD round state store 扩展与 checker preview token store
  - 不负责：承接 AWD 业务编排或 validation owner
- `contest/runtime`
  - 负责：统一注入 `AWDRoundStateStore` 与 `AWDCheckerPreviewTokenStore`
  - 不负责：把 Redis client 暴露回 AWD application

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-13-contest-awd-runtime-state-and-preview-token-store-phase5-slice8-implementation-plan.md`
- Add: `.harness/reuse-decisions/contest-awd-runtime-state-and-preview-token-store-phase5-slice8.md`
- Add: `code/backend/internal/module/contest/infrastructure/awd_checker_preview_token_store.go`
- Add: `code/backend/internal/module/contest/ports/awd_checker_preview_token_context_contract_test.go`
- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/module/contest/ports/awd_round_state_context_contract_test.go`
- Modify: `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_current_round_fallback_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_flag_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_status_cache.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_checker_preview_token_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_attack_log_response_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_upsert_response_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

- [x] Slice 1: 扩展 AWD state stores 并切走 AWDService / ContestAWDServiceService 的 Redis direct access
  - 目标：AWD application 不再直接持有 Redis client；preview token 与 round state 行为保持一致
  - 验证：
    - `cd code/backend && go test ./internal/module/contest/application/commands -run 'AWDService|ContestAWDServiceService' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`
  - Review focus：token 匹配/消费语义、current round fallback、flag fallback、service status 更新是否保持不变

- [x] Slice 2: 删除剩余 contest AWD Redis allowlist 并同步文档
  - 目标：删除本次实际收口的 AWD Redis allowlist；更新 phase 5 当前事实
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除本次 AWD surface 实际收口的例外；文档不误写成 contest 已无任何 Redis 依赖之外的事实

## Risks

- 如果 checker preview token 匹配后删除语义回归，会导致错误消费或 token 失效
- 如果 current round fallback 与 round flag fallback 不一致，AWD flag 提交可能回归
- 如果 service status cache 更新丢失“只在当前轮次时写入”的约束，会污染 runtime state

## Verification Plan

1. `cd code/backend && go test ./internal/module/contest/application/commands -run 'AWDService|ContestAWDServiceService' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：AWD application 继续持有轮次/flag/preview validation 业务编排，Redis 状态细节落回 infrastructure
- reuse point 明确：优先复用已有 `AWDRoundStateStore`，preview token 单独抽成窄 store，而不是让两个 service 各自复制 Redis helper
- 这刀收口后，contest phase 5 在 AWD / submission / scoreboard 这三条剩余 Redis surface 上都能完成 application concrete allowlist 清理
