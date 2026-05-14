# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/contest/application/jobs/awd_check_cache_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_runtime.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_flag_support.go`
- `code/backend/internal/module/contest/infrastructure/awd_command_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_query_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_manager_adapter.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/awd.go`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/awd_command_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_manager_adapter.go`
- `code/backend/internal/module/contest/infrastructure/awd_query_repository.go`

## Decision

refactor_existing

## Reason

这次 debt 不是 AWD jobs 需要新的业务能力，而是 `AWDRoundUpdater` 仍直接知道 round lookup / active-round materialize 的 concrete not-found 语义。沿用 phase5 已验证模式，最小正确做法是：

- 继续复用已有 `contestports.ErrContestAWDRoundNotFound`
- 在 `contest/infrastructure` 新增 job-only adapter，包装 `AWDRoundUpdater` 使用的 repo 面，把 `FindRunningRound` / `FindRoundByNumber` 的 concrete not-found 收口成 contest sentinel
- `AWDRoundUpdater.EnsureActiveRoundMaterialized` 直接返回 contest sentinel，不再返回 `gorm.ErrRecordNotFound`
- `contest/runtime/module.go` 只给 job 路径注入这个 adapted repo，不改 query / command 路径

这样可以把 touched surface 上的结构债一并收口：当前 `buildAWDHandler` 直接把 raw `AWDRepository` 注给 `AWDRoundUpdater`，如果只改 job 文件不改 wiring，会留下同一位置的未收口 concrete dependency。

## Files to modify

- `.harness/reuse-decisions/contest-awd-job-round-lookup-contract-phase5-slice39.md`
- `docs/plan/impl-plan/2026-05-14-contest-awd-job-round-lookup-contract-phase5-slice39-implementation-plan.md`
- `code/backend/internal/module/contest/infrastructure/awd_job_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_job_repository_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_check_cache_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_runtime.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_contract_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_testsupport_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## After implementation

- `awd_check_cache_support.go -> gorm.io/gorm` 收口到 adapted repo + `ErrContestAWDRoundNotFound`
- `awd_round_flag_lookup_support.go -> gorm.io/gorm` 收口到 adapted repo + `ErrContestAWDRoundNotFound`
- `awd_round_runtime.go -> gorm.io/gorm` 改为直接返回 `ErrContestAWDRoundNotFound`
- `contest/runtime/module.go` 只在 `AWDRoundUpdater` wiring 上引入 adapted repo
- HTTP checker / probe / target client 相关 `net/http` allowlist 保持给下一刀处理
