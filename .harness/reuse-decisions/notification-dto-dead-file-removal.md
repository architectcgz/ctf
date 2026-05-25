# Reuse Decision

## Change type

cleanup / ops / dto / dead-code

## Existing code searched

- `code/backend/internal/dto/notification.go`
- `code/backend/internal/module/ops/api/http/notification_types.go`
- `code/backend/internal/module/ops/application/commands/notification_output.go`
- `code/backend/internal/module/ops/application/queries/notification_output.go`

## Similar implementations found

- `ops/api/http/notification_types.go` 已承接 HTTP request / response 类型
- `ops/application/commands/notification_output.go` 已承接 command output 和 audience 常量
- `ops/application/queries/notification_output.go` 已承接 query output

## Decision

remove_dead_code

## Reason

`internal/dto/notification.go` 的类型已经全部被 ops 模块内 owner 文件替代，仓库里没有剩余消费方。继续保留只会制造全局 DTO 仍是事实源的假象。这一刀不迁移行为，只删除无引用空壳。

## Files to modify

- `.harness/reuse-decisions/notification-dto-dead-file-removal.md`
- `docs/plan/archive/impl-plan/2026-05-18-notification-dto-dead-file-removal-implementation-plan.md`
- `code/backend/internal/dto/notification.go`

## After implementation

- `internal/dto/notification.go` 删除
- ops 通知相关代码继续只依赖模块内 owner 类型
