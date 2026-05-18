# ops risk DTO 收口实现方案

## Objective

把 `internal/dto/cheat_detection.go` 中属于 `ops` 风险查询的类型拆回 owning module：

- `ops/application/queries`：`CheatDetectionResp` 及其子类型

## Non-goals

- 不在这一刀里处理 `notification.go`
- 不改变作弊检测接口的字段语义和排序逻辑

## Inputs

- `code/backend/internal/dto/cheat_detection.go`
- `code/backend/internal/module/ops/api/http/risk_handler.go`
- `code/backend/internal/module/ops/application/queries/risk_service.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Task slices

1. `ops/application/queries`
   - 新增风险检测输出类型
   - risk service 改依赖新 owner

2. `ops/api/http` 与 app tests
   - handler 和 full router 集成测试改依赖新 owner

3. cleanup
   - 删除 `internal/dto/cheat_detection.go`
   - 跑 ops 与受影响 app 用例

## Expected changes

- `code/backend/internal/dto/cheat_detection.go`
- `code/backend/internal/module/ops/api/http/**`
- `code/backend/internal/module/ops/application/queries/**`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Validation

- `go test ./internal/module/ops/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_AdminRiskDashboardAndImageStateMatrix' -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `ops/application/queries` 是否成为 risk query output 唯一 owner
- handler / app 测试是否彻底脱离 `dto.CheatDetection*`
