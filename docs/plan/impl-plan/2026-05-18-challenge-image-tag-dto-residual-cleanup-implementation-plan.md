# challenge image/tag DTO 残留清理实现方案

## Objective

清掉 `internal/dto` 中已经完成 owner 迁移但仍遗留的 `image.go`、`tag.go`：

- app 集成测试改为使用 `challenge/api/http.ImageResp`
- 删除零消费的全局 image/tag DTO 文件

## Non-goals

- 不在这一刀处理 `challenge.go`、`topology.go`、`awd_challenge*.go`、`common.go`
- 不改 image/tag 接口字段、状态码、分页语义
- 不改 challenge 模块现有 contracts / mapper 结构

## Inputs

- `code/backend/internal/dto/{image.go,tag.go}`
- `code/backend/internal/module/challenge/api/http/{challenge_request_types.go,challenge_response_types.go}`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `.harness/reuse-decisions/challenge-image-tag-dto-residual-cleanup.md`

## Task slices

1. app 测试收口
   - `full_router_state_matrix_integration_test.go` 改为直接解码 `challenge/api/http.ImageResp`

2. cleanup
   - 删除 `internal/dto/image.go`
   - 删除 `internal/dto/tag.go`

3. verification
   - 跑受影响 app 用例、challenge 模块测试、全局 mapper guardrail 和 consistency check

## Expected changes

- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/image.go`
- `code/backend/internal/dto/tag.go`

## Validation

- `go test ./internal/app -run 'TestFullRouter_AdminOpsAndNotificationStateMatrix|TestFullRouter_AdminImagesCapsOversizedPageSize' -count=1`
- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- app 测试是否彻底脱离 `dto.ImageResp`
- `internal/dto/image.go`、`internal/dto/tag.go` 删除后是否还存在引用遗漏
- 本刀是否只做残留清理，没有把 image/tag DTO 再次放回全局中间层
