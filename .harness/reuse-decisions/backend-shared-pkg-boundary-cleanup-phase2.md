# Reuse Decision

## Change type
transport / handler / middleware / architecture guardrail

## Existing code searched

- `code/backend/pkg/response/`
- `code/backend/internal/middleware/`
- `code/backend/internal/handler/`
- `code/backend/internal/app/`
- `code/backend/internal/module/*/api/http/`
- `code/backend/internal/infrastructure/`

## Similar implementations found

- `code/backend/internal/infrastructure/ratelimit/ratelimit.go`
- `code/backend/internal/module/*/api/http/*.go`
- `code/backend/internal/handler/health/handler.go`
- `code/backend/internal/middleware/recovery.go`

## Decision
refactor_existing

## Reason

`pkg/response` 不是对外公共库，也不是某个业务模块私有能力。它只服务于仓库内部的 Gin handler、middleware、router 和健康检查入口，本质上是共享 HTTP transport helper。最小正确方案不是把成功/失败响应逻辑复制到各模块 `api/http`，也不是塞进 `internal/infrastructure`，而是沿用现有函数 API 与 envelope 结构，整体迁到明确的内部 transport 包 `internal/httpresponse`，统一让所有 HTTP 入口继续复用同一套 helper。

## Files to modify

- `code/backend/pkg/response/response.go`
- `code/backend/pkg/response/response_test.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/handler/health/handler.go`
- `code/backend/internal/middleware/{auth,parse_id,rbac,ratelimit,recovery}.go`
- `code/backend/internal/module/*/api/http/*.go`
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
- `code/backend/internal/module/runtime/architecture_test.go`
- `code/backend/internal/module/contest/api/http/awd_round_check_handler.go`
- `code/backend/internal/module/contest/api/http/challenge_add_handler.go`
- `code/backend/internal/module/contest/api/http/challenge_manage_handler.go`
- `code/backend/internal/module/contest/api/http/contest_command_handler.go`
- `code/backend/internal/module/contest/api/http/participation_announcement_handler.go`
- `code/backend/internal/module/contest/api/http/participation_registration_handler.go`
- `code/backend/internal/module/contest/api/http/scoreboard_admin_handler.go`
- `code/backend/internal/module/contest/api/http/team_create_join_handler.go`
- `code/backend/internal/module/contest/api/http/team_manage_handler.go`
- `code/backend/internal/module/ops/api/http/notification_handler.go`

## After implementation

- 若后续 `pkg/websocket` 也按“内部 transport / infra owner”规则迁移，可把这轮 shared pkg 清理规则统一沉淀到长期 reuse 索引。
