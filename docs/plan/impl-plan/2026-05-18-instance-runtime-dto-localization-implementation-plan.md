# instance / runtime DTO 收口实现方案

## Objective

把 `internal/dto/instance.go` 中属于 `instance / runtime` 的类型拆回 owning module：

- `instance/contracts`：实例 owner 对外返回值与 AWD defense workbench 契约
- `runtime/api/http`：仅 runtime HTTP endpoint 使用的访问响应

## Non-goals

- 不在这一刀里彻底收掉 `internal/dto/contest_awd_instance.go`
- 不改实例、AWD、代理访问的 JSON 字段语义
- 不扩到 `challenge / image / tag / topology / notification` 等其他 DTO 文件

## Inputs

- `code/backend/internal/dto/instance.go`
- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/module/instance/**/*`
- `code/backend/internal/module/runtime/**/*`
- `code/backend/internal/module/practice/**/*instance*.go`
- `code/backend/internal/app/composition/runtime_http_service_adapter.go`
- `code/backend/internal/testutil/runtimeadapters/http_service.go`

## Task slices

1. `instance/contracts`
   - 新增实例 owner 输出类型与 AWD defense workbench 契约类型
   - `instance` application / contracts 不再依赖全局 `dto.Instance*`

2. `runtime/api/http`
   - 新增 runtime 本地访问响应类型
   - runtime handler / tests 不再依赖全局 `dto.InstanceAccessResp` / `dto.AWDDefenseSSHAccessResp`

3. consumers
   - `practice`、`app composition`、`testutil`、相关测试改为依赖新 owner
   - `contest_awd_instance.go` 若仍存在，只引用新 owner 类型

4. cleanup
   - 删除 `internal/dto/instance.go`
   - 跑 mapper 生成与受影响测试

## Expected changes

- `code/backend/internal/dto/instance.go`
- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/module/instance/**`
- `code/backend/internal/module/runtime/**`
- `code/backend/internal/module/practice/**`
- `code/backend/internal/app/composition/runtime_http_service_adapter.go`
- `code/backend/internal/testutil/runtimeadapters/http_service.go`
- `code/backend/internal/app/**test.go`

## Validation

- `go generate ./internal/module/instance/... ./internal/module/runtime/... ./internal/module/practice/...`
- `go test ./internal/module/instance/... -count=1`
- `go test ./internal/module/runtime/... -count=1`
- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge|TestFullRouter_ListInstancesMatchesContract|TestFullRouter_StudentChallengeLifecycleStateMatrix' -count=1`

## Review focus

- `instance` owner 是否成为 `InstanceResp / InstanceInfo / defense workbench` 唯一 owner
- `runtime/api/http` 是否只保留 HTTP 展示层响应
- `practice` 与 `app` 是否没有重新把全局 `dto.Instance*` 漏回去
