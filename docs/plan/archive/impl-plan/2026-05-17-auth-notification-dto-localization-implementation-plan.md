# auth / notification DTO 模块内化实现方案

## Objective

把 `internal/dto/auth.go`、`internal/dto/notification.go` 中由单模块独占的 request/response 类型收回 owning module，保持外部 HTTP 契约不变：

- `auth`：认证、CAS、WS ticket 的 HTTP request/response 与 command/query output
- `ops(notification)`：通知查询、管理员发布通知的 HTTP request/response 与 command/query output

## Non-goals

- 不改 `auth`、`ops` 的路由、权限和 JSON 字段
- 不改通知事件总线、WebSocket topic 和推送时机
- 不处理 `challenge / contest / instance / awd` 的剩余全局 DTO（独立下一批）

## Inputs

- `docs/architecture/backend/04-api-design.md`
- `docs/plan/impl-plan/2026-05-17-identity-admin-user-http-dto-localization-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-17-ops-dashboard-http-dto-localization-implementation-plan.md`
- `code/backend/internal/module/auth/**/*`
- `code/backend/internal/module/ops/**/*notification*.go`
- `code/backend/internal/module/ops/api/http/request_mapper*.go`
- `code/backend/internal/dto/auth.go`
- `code/backend/internal/dto/notification.go`

## Task Slices

1. `auth` output 内化
   - 在 `auth/application/commands` 新增 login output 类型
   - 在 `auth/application/queries` 新增 CAS output 类型
   - command/query service 与 mapper 改为仅依赖模块内类型

2. `auth` HTTP DTO 内化
   - 在 `auth/api/http` 新增 request/response 类型
   - handler 与 request/response mapper 改为仅依赖本地 HTTP DTO 和 `auth` application output

3. `ops(notification)` output 内化
   - 在 `ops/application/commands`、`ops/application/queries` 新增 notification output/query 类型
   - 迁移 audience 常量 owner 到 `ops/application/commands`

4. `ops(notification)` HTTP DTO 内化
   - 在 `ops/api/http` 新增 notification request/response/query 类型
   - handler 与 request mapper 只依赖模块内类型

5. 测试与全局 DTO 清理
   - 调整集成测试、notification 相关单测解码类型
   - 删除 `internal/dto/auth.go`、`internal/dto/notification.go`（确认无剩余引用）

## Expected Changes

- `code/backend/internal/module/auth/**`
- `code/backend/internal/module/ops/**/*notification*.go`
- `code/backend/internal/module/ops/api/http/request_mapper*.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/auth.go`
- `code/backend/internal/dto/notification.go`
- `docs/architecture/backend/04-api-design.md`

## Validation

- `go generate ./internal/module/auth/... ./internal/module/ops/...`
- `go test ./internal/module/auth/... -count=1`
- `go test ./internal/module/ops/... -count=1`
- `go test ./internal/app -run TestFullRouter_AdminOpsAndNotificationStateMatrix -count=1`

## Review Focus

- `auth` 与 `ops(notification)` 是否成为对应 DTO 的唯一 owner
- application 边界是否不再依赖 `internal/dto`
- 外部 HTTP 契约与 WebSocket payload 字段是否保持兼容
