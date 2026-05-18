# practice contest AWD instance DTO 收口实现方案

## Objective

把 `internal/dto/contest_awd_instance.go` 中属于 `practice` admin AWD 实例编排的类型拆回 owning module：

- `practice/api/http`：仅 handler 绑定 JSON 请求使用的 request DTO
- `practice/application/commands`：admin AWD orchestration / control / prewarm 输出类型

## Non-goals

- 不在这一刀里处理 `score.go`、`challenge.go`、`image.go` 等其他全局 DTO 文件
- 不改变 admin AWD 实例编排和控制接口的 JSON 字段语义
- 不扩到 contest / challenge 其它 owner 归属调整

## Inputs

- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice/application/commands/*.go`
- `code/backend/internal/app/full_router_integration_test.go`

## Task slices

1. `practice/application/commands`
   - 新增 admin AWD orchestration / control / prewarm 输出类型
   - command service / mapper / tests 不再依赖 `dto.AdminAWD*`

2. `practice/api/http`
   - 新增 admin AWD handler 请求类型
   - handler 只依赖本地 request DTO 和 command output DTO

3. consumers
   - `full_router` app 测试改为依赖 `practice/application/commands` 输出类型

4. cleanup
   - 删除 `internal/dto/contest_awd_instance.go`
   - 跑 practice mapper 生成与受影响测试

## Expected changes

- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/module/practice/api/http/**`
- `code/backend/internal/module/practice/application/commands/**`
- `code/backend/internal/app/full_router_integration_test.go`

## Validation

- `go generate ./internal/module/practice/...`
- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `go test ./internal/app -run 'TestFullRouter_AdminContestAWDScopeControlsAffectLifecycleAndOrchestration' -count=1`

## Review focus

- request DTO 是否只留在 `practice/api/http`
- admin AWD orchestration / control / prewarm 输出是否由 `practice/application/commands` 成为唯一 owner
- `full_router` 与 practice 测试是否已经脱离 `dto.AdminAWD*`
