# notification DTO 空壳删除实现方案

## Objective

删除已经无消费方的 `internal/dto/notification.go`。

## Non-goals

- 不改 ops 通知模块现有 request / response owner
- 不调整通知发布、查询或 websocket 行为

## Inputs

- `code/backend/internal/dto/notification.go`
- `code/backend/internal/module/ops/**`

## Task slices

1. 确认无消费方
2. 删除 `internal/dto/notification.go`
3. 跑最小编译与相关测试

## Expected changes

- `code/backend/internal/dto/notification.go`

## Validation

- `go test ./internal/dto ./internal/module/ops/... -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- 是否确实没有残留 `dto.Notification*` / `dto.AdminNotification*` 引用
- 删除后 ops 通知模块是否仍完全依赖模块内 owner
