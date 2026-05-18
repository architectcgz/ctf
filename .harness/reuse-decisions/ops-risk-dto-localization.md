# Reuse Decision

## Change type

contract / query / ops / risk

## Existing code searched

- `code/backend/internal/dto/cheat_detection.go`
- `code/backend/internal/module/ops/api/http/risk_handler.go`
- `code/backend/internal/module/ops/application/queries/risk_service.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Similar implementations found

- `ops/application/queries/notification_output.go` 已承接 ops query request / response 类型
- `ops/application/queries/audit_output.go` 刚完成 audit query contract 收口

## Decision

refactor_existing

## Reason

`CheatDetectionResp` 及其子类型当前只被 ops 风险查询 service、risk handler 和 app 集成测试解码使用，owner 已经很单一。最小正确方案是把这组类型收回 `ops/application/queries`，保持 ops query 输出都由查询层自己持有，不再依赖全局 `dto`。

## Files to modify

- `.harness/reuse-decisions/ops-risk-dto-localization.md`
- `docs/plan/impl-plan/2026-05-18-ops-risk-dto-localization-implementation-plan.md`
- `code/backend/internal/dto/cheat_detection.go`
- `code/backend/internal/module/ops/api/http/risk_handler.go`
- `code/backend/internal/module/ops/application/queries/risk_output.go`
- `code/backend/internal/module/ops/application/queries/risk_service.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## After implementation

- ops risk handler / query service / app 测试不再引用 `dto.CheatDetection*`
- `internal/dto/cheat_detection.go` 删除
