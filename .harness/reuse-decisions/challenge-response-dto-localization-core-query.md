# Reuse Decision

## Change type

api / response mapper / dto localization

## Existing code searched

- `code/backend/internal/module/challenge/api/http/handler.go`
- `code/backend/internal/module/challenge/api/http/response_mapper*.go`
- `code/backend/internal/dto/{challenge.go,challenge_import.go,common.go}`
- `docs/plan/archive/impl-plan/2026-05-17-challenge-response-dto-localization-topology-awd-implementation-plan.md`

## Similar implementations found

- `challenge` 的 image/tag/flag/topology/awd 响应已采用 mapper 收口
- `contest` request DTO 收口沿用“handler 绑定/输出只走模块内类型”模式

## Decision

refactor_existing

## Reason

主 `challenge` handler 里的列表/详情/创建结果/导入提交仍直接透传全局 `dto`。本切片把这组查询面和提交结果收口到 `challenge/api/http` 本地 response DTO，保持外部契约不变。

## Files to modify

- `.harness/reuse-decisions/challenge-response-dto-localization-core-query.md`
- `docs/plan/archive/impl-plan/2026-05-17-challenge-response-dto-localization-core-query-implementation-plan.md`
- `code/backend/internal/module/challenge/api/http/challenge_response_types.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
- `code/backend/internal/module/challenge/api/http/response_mapper_gen.go`
- `code/backend/internal/module/challenge/api/http/handler.go`
