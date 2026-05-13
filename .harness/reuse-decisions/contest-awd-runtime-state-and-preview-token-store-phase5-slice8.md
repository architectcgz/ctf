# Reuse Decision

## Change type

command / port / infrastructure / runtime

## Existing code searched

- `code/backend/internal/module/contest/application/commands/awd_*.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_*.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/infrastructure/`
- `docs/plan/impl-plan/2026-05-13-contest-submission-rate-limit-store-phase5-slice7-implementation-plan.md`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`
- `code/backend/internal/module/contest/infrastructure/scoreboard_state_store.go`
- `code/backend/internal/module/contest/application/commands/awd_checker_preview_token_support.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`

## Decision

refactor_existing

## Reason

当前 debt 不是 AWD 缺少状态能力，而是 current round fallback、round flag、service status 与 checker preview token 的 Redis 细节同时散落在 `AWDService` 和 `ContestAWDServiceService`。最小正确方案不是继续保留 application helper 直连 Redis，也不是分裂成两套平行 token helper，而是复用 phase5 既有 state-store 思路：扩展 `AWDRoundStateStore` 承接 round/service status 状态，再新增窄 `AWDCheckerPreviewTokenStore` 承接 preview token 持久化，供两个 AWD command surface 共用。

## Files to modify

- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/ports/awd_round_state_context_contract_test.go`
- `code/backend/internal/module/contest/ports/awd_checker_preview_token_context_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`
- `code/backend/internal/module/contest/infrastructure/awd_checker_preview_token_store.go`
- `code/backend/internal/module/contest/application/commands/awd_service.go`
- `code/backend/internal/module/contest/application/commands/awd_current_round_fallback_support.go`
- `code/backend/internal/module/contest/application/commands/awd_flag_support.go`
- `code/backend/internal/module/contest/application/commands/awd_status_cache.go`
- `code/backend/internal/module/contest/application/commands/awd_checker_preview_token_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_log_response_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service_upsert_response_support.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-contest-awd-runtime-state-and-preview-token-store-phase5-slice8-implementation-plan.md`

## After implementation

- 如果以后 AWD 还要增加更多短期 runtime state，优先继续落到现有 `AWDRoundStateStore` 或专用窄 store，而不是把 Redis 细节拉回 application
- contest phase 5 完成后，后续 allowlist 收缩应转向其他模块，而不是继续在 contest application 留 Redis 例外
