# Reuse Decision

## Change type

cleanup / auth / dto / dead-code

## Existing code searched

- `code/backend/internal/dto/auth.go`
- `code/backend/internal/module/auth/**/*.go`
- `code/backend/internal/app/**/*.go`

## Similar implementations found

- `auth/api/http/auth_types.go` 已承接 HTTP request / response 类型
- `auth/application/commands/login_output.go` 已承接登录输出类型
- `auth/application/queries/cas_output.go` 已承接 CAS 查询输出类型

## Decision

remove_dead_code

## Reason

`internal/dto/auth.go` 的类型已经全部被 auth 模块内的 owner 文件替代，仓库里没有剩余消费方。继续保留只会制造“全局 DTO 仍是事实源”的错觉。这一刀不迁移行为，只删除无引用空壳。

## Files to modify

- `.harness/reuse-decisions/auth-dto-dead-file-removal.md`
- `docs/plan/archive/impl-plan/2026-05-18-auth-dto-dead-file-removal-implementation-plan.md`
- `code/backend/internal/dto/auth.go`

## After implementation

- `internal/dto/auth.go` 删除
- auth 相关模块继续只依赖模块内 owner 类型
