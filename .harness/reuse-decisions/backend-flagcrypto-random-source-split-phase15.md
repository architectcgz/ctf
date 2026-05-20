# Reuse Decision

## Change type
shared flag crypto boundary cleanup

## Existing code searched

- `code/backend/internal/shared/flagcrypto/flag.go`
- `code/backend/internal/config/config.go`
- `code/backend/internal/module/challenge/**/*.go`
- `code/backend/internal/module/practice/**/*.go`
- `code/backend/internal/module/contest/**/*.go`
- `code/backend/internal/platform/**/*.go`

## Similar implementations found

- `code/backend/internal/middleware/request_id.go`
- `code/backend/internal/module/auth/infrastructure/token_service.go`
- `code/backend/internal/module/instance/application/queries/proxy_ticket_service.go`
- `code/backend/internal/module/contest/application/commands/team_support.go`

## Decision
refactor_existing

## Reason

`internal/shared/flagcrypto` 当前同时承载两类不同 owner 的能力：

- 真正的跨模块 flag 算法：动态 flag 生成、静态 flag 哈希、常数时间比较
- 通用随机串生成：salt / nonce / config secret

前者被 `challenge / practice / contest` 共同复用，shared 语义成立；后者只是平台级技术能力，还让 `config` 反向依赖了 `flagcrypto`，使包语义变脏。

因此这次不把 `flagcrypto` 收回某个模块，而是把随机串生成拆到明确的平台技术包 `internal/platform/randomstring`，让 `flagcrypto` 只保留 flag 算法。

## Files to modify

- `code/backend/internal/shared/flagcrypto/flag.go`
- `code/backend/internal/shared/flagcrypto/flag_test.go`
- `code/backend/internal/platform/randomstring/random.go`
- `code/backend/internal/platform/randomstring/random_test.go`
- `code/backend/internal/config/config.go`
- `code/backend/internal/module/challenge/application/commands/flag_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_self_check_test.go`

## After implementation

- `internal/shared/flagcrypto` 只保留 flag 算法相关能力
- `config` 不再依赖 `flagcrypto`
- salt / nonce / secret 随机串生成收敛到 `internal/platform/randomstring`
