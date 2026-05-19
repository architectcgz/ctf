# Reuse Decision

## Change type
transport / infrastructure / runtime / ports / handler

## Existing code searched

- `code/backend/pkg/websocket/`
- `code/backend/internal/module/ops/`
- `code/backend/internal/module/contest/`
- `code/backend/internal/app/composition/`
- `code/backend/internal/infrastructure/`

## Similar implementations found

- `code/backend/internal/httpresponse/response.go`
- `code/backend/internal/infrastructure/ratelimit/ratelimit.go`
- `code/backend/internal/module/contest/api/http/realtime_handler.go`
- `code/backend/internal/module/ops/api/http/notification_handler.go`

## Decision
refactor_existing

## Reason

`pkg/websocket` 同时承载了两类 owner：

- 共享 WebSocket 消息契约：`Envelope`
- 具体连接管理实现：`Manager`

如果整包直接迁到 `internal/infrastructure/websocket`，`ops/contest` 的 ports 和 application 还会继续反向依赖 infrastructure；如果整包只迁到 `internal/websocket`，又会让应用层继续耦合到实现细节和配置依赖。最小正确方案是拆成两层：

- `internal/websocket` 只保留共享消息契约
- `internal/infrastructure/websocket` 持有 `Manager` 实现

同时把 `ops/api/http/notification_handler.go` 收成窄接口依赖，避免 API 层知道具体 `Manager` 类型。

## Files to modify

- `code/backend/pkg/websocket/manager.go`
- `code/backend/pkg/websocket/manager_test.go`
- `code/backend/internal/websocket/*.go`
- `code/backend/internal/infrastructure/websocket/*.go`
- `code/backend/internal/module/ops/ports/notification.go`
- `code/backend/internal/module/contest/ports/realtime.go`
- `code/backend/internal/module/ops/application/commands/contest_realtime_service.go`
- `code/backend/internal/module/ops/application/commands/contest_realtime_service_test.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- `code/backend/internal/module/ops/api/http/notification_handler.go`
- `code/backend/internal/module/ops/api/http/notification_http_integration_test.go`
- `code/backend/internal/module/ops/runtime/module.go`
- `code/backend/internal/module/ops/architecture_test.go`
- `code/backend/internal/app/composition/ops_module.go`

## After implementation

- 如果后续还要清 `pkg/crypto` 或类似“契约 + 实现混包”场景，可以沿用这套“共享 transport contract + internal infrastructure implementation”的拆法。
