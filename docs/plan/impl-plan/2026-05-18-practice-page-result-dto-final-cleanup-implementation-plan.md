# practice PageResult DTO 最终清理实现方案

## Objective

收掉 `internal/dto` 的最后遗留：

- `common.go` 的 `PageResult[T]` 收回 `practice/contracts`
- `challenge_test.go` 的 `ConfigureFlagReq` 绑定测试迁回 `challenge/api/http`
- 删除 `internal/dto` 目录剩余文件

## Non-goals

- 不改 manual review 列表接口 JSON 字段
- 不改 practice manual review 查询逻辑、分页默认值和权限判断
- 不改 challenge flag 配置接口行为

## Inputs

- `code/backend/internal/dto/{common.go,common_test.go,challenge_test.go}`
- `code/backend/internal/module/practice/{contracts,api/http,application/commands}`
- `code/backend/internal/module/challenge/api/http/challenge_request_types.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

## Task slices

1. `practice/contracts`
   - 新增 `PageResult[T]`
   - manual review service / handler 改依赖 `practicecontracts.PageResult`

2. app / challenge tests
   - `practice_flow_integration_test.go` 改为解码 `practicecontracts.PageResult`
   - 新增 `challenge/api/http` 侧的 `ConfigureFlagReq` 绑定测试

3. cleanup
   - 删除 `internal/dto/common.go`
   - 删除 `internal/dto/common_test.go`
   - 删除 `internal/dto/challenge_test.go`

## Expected changes

- `code/backend/internal/module/practice/contracts/page_result.go`
- `code/backend/internal/module/practice/contracts/page_result_test.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- `code/backend/internal/module/challenge/api/http/configure_flag_req_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/dto/common.go`
- `code/backend/internal/dto/common_test.go`
- `code/backend/internal/dto/challenge_test.go`

## Validation

- `go test ./internal/app -run 'TestPracticeFlow_.*' -count=1`
- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module/challenge/api/http -run TestConfigureFlagReqRejectsSharedProofFlagType -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `practice/contracts.PageResult` 是否成为 manual review 分页链路唯一 owner
- `practice_flow` 和 practice handler 是否完全脱离 `internal/dto`
- 删除 `internal/dto` 后 challenge 绑定测试是否仍覆盖原约束
