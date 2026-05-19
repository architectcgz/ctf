# Reuse Decision

## Change type
shared kernel / domain support / config / command

## Existing code searched

- `code/backend/pkg/crypto/`
- `code/backend/internal/shared/`
- `code/backend/internal/module/challenge/`
- `code/backend/internal/module/practice/`
- `code/backend/internal/module/contest/`
- `code/backend/internal/config/`
- `code/backend/cmd/`

## Similar implementations found

- `code/backend/internal/shared/taxonomy/`
- `code/backend/internal/shared/mapperhelper/`
- `code/backend/internal/shared/mapperutil/`

## Decision
refactor_existing

## Reason

`pkg/crypto/flag.go` 不是通用密码学工具箱，而是项目自己的 flag 生成 / 校验算法能力，当前被 challenge、practice、contest、config 和命令入口共同依赖。它不属于单个业务模块，也不应该落到 `internal/infrastructure`。最小正确方案是把它收回共享内核层 `internal/shared/flagcrypto`，继续保留纯函数 API，不引入新的服务抽象。

## Files to modify

- `code/backend/pkg/crypto/flag.go`
- `code/backend/pkg/crypto/flag_test.go`
- `code/backend/internal/config/config.go`
- `code/backend/internal/module/challenge/application/{commands,queries}/*.go`
- `code/backend/internal/module/practice/application/commands/*.go`
- `code/backend/internal/module/contest/domain/awd_flag_support.go`
- `code/backend/internal/module/contest/application/commands/awd_attack_submit_support.go`
- `code/backend/internal/module/contest/testsupport/fixtures.go`
- `code/backend/cmd/seed-demo-challenges/main.go`
- `code/backend/cmd/import-challenge-packs/main.go`

## After implementation

- `errcode` 若后续继续收口，可以参考这刀的 owner 判断方式：先确认它是共享内核、transport contract，还是模块私有错误语义，再决定落点。
