# Reuse Decision

## Change type
service / port / repository / runtime

## Existing code searched
- `code/backend/internal/module/contest/application/commands`
- `code/backend/internal/module/contest/application/jobs`
- `code/backend/internal/module/contest/infrastructure`
- `code/backend/internal/module/contest/ports`
- `code/backend/internal/module/contest/runtime`

## Similar implementations found
- `code/backend/internal/module/contest/application/jobs/awd_round_flag_sync.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`
- `code/backend/internal/module/contest/infrastructure/awd_docker_flag_injector.go`
- `code/backend/internal/module/contest/runtime/module.go`

## Decision
extend_existing

## Reason
当前仓库已经具备整轮 flag 同步能力，包括 Redis round state 写入、Docker 容器 `/flag` 注入、以及 runtime 装配点。本次不新建独立轮换模块，而是在现有 `AWDService`、`AWDRoundStateStore` 和 `AWDFlagInjector` 基础上扩展单条即时轮换与原子 claim 能力，保持 owner 清晰并复用现有 flag 注入路径。

## Files to modify
- `code/backend/internal/module/contest/application/commands/awd_service.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_submit_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_submit_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_flag_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
- `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`
- `code/backend/internal/module/contest/ports/contest.go`
- `code/backend/internal/module/contest/runtime/module.go`

## After implementation
- 如形成稳定模式，再决定是否补 `harness/reuse/history.md` 或 `harness/reuse/index.yaml`。
