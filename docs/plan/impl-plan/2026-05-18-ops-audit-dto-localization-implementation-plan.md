# ops audit DTO 收口实现方案

## Objective

把 `internal/dto/audit.go` 中属于 `ops` 审计查询的类型拆回 owning module：

- `ops/application/queries`：`AuditLogQuery`、`AuditLogItem`

## Non-goals

- 不在这一刀里处理 `cheat_detection.go`、`notification.go`
- 不改变审计查询的参数语义、分页语义和 JSON 字段语义

## Inputs

- `code/backend/internal/dto/audit.go`
- `code/backend/internal/module/ops/api/http/audit_handler.go`
- `code/backend/internal/module/ops/application/queries/audit_service.go`
- `code/backend/internal/module/ops/application/queries/audit_service_test.go`

## Task slices

1. `ops/application/queries`
   - 新增审计查询 request / response 类型
   - query service 和 tests 改依赖新 owner

2. `ops/api/http`
   - handler 改为绑定本地 query owner

3. cleanup
   - 删除 `internal/dto/audit.go`
   - 跑 ops 模块相关测试

## Expected changes

- `code/backend/internal/dto/audit.go`
- `code/backend/internal/module/ops/api/http/**`
- `code/backend/internal/module/ops/application/queries/**`

## Validation

- `go test ./internal/module/ops/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `ops/application/queries` 是否成为 audit query contract 唯一 owner
- handler / service / tests 是否彻底脱离全局 `dto.AuditLog*`
