# Reuse Decision

## Change type

backend / http response dto localization

## Existing code searched

- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/*handler*.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
- `code/backend/internal/module/ops/api/http/notification_response_mapper.go`

## Similar implementations found

- `challenge/api/http` 已将 HTTP response DTO 和 response mapper 收回模块边界
- `ops/api/http` 已把 notification response 映射独立为专门的 response mapper
- `contest` 当前 request DTO 已在 `api/http` 本地化，但 response 仍历史性挂在 request mapper 上

## Decision

refactor_existing

## Reason

当前 `contest/api/http` 的真实问题不是“少几个 dto import”，而是 request / response owner 混在同一 mapper，且 response 仍落到全局 `internal/dto`。最小正确修复是沿已有模块化模式补齐本地 response DTO 和独立 response mapper，而不是继续在 request mapper 上叠历史兼容层。

## Files to modify

- `.harness/reuse-decisions/contest-http-response-dto-localization.md`
- `docs/plan/impl-plan/2026-05-18-contest-response-dto-localization-implementation-plan.md`
- `code/backend/internal/app/challenge_import_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/awd.go`
- `code/backend/internal/dto/contest.go`
- `code/backend/internal/dto/contest_awd_service.go`
- `code/backend/internal/dto/contest_awd_workspace.go`
- `code/backend/internal/dto/contest_challenge.go`
- `code/backend/internal/dto/team.go`
- `code/backend/internal/middleware/awd_readiness_audit.go`
- `code/backend/internal/module/contest/api/http/awd_attack_handler.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_handler.go`
- `code/backend/internal/module/contest/api/http/awd_round_manage_handler.go`
- `code/backend/internal/module/contest/api/http/awd_round_summary_handler.go`
- `code/backend/internal/module/contest/api/http/awd_service_handler.go`
- `code/backend/internal/module/contest/api/http/awd_traffic_handler.go`
- `code/backend/internal/module/contest/api/http/awd_workspace_handler.go`
- `code/backend/internal/module/contest/api/http/challenge_query_handler.go`
- `code/backend/internal/module/contest/api/http/contest_response_types.go`
- `code/backend/internal/module/contest/api/http/participation_query_handler.go`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_assign.go`
- `code/backend/internal/module/contest/api/http/request_mapper_awd_service_support.go`
- `code/backend/internal/module/contest/api/http/request_mapper_contest_support.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/response_mapper.go`
- `code/backend/internal/module/contest/api/http/response_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/scoreboard_query_handler.go`
- `code/backend/internal/module/contest/api/http/team_query_handler.go`
- `code/backend/internal/module/contest/architecture_test.go`
