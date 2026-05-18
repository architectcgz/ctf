# auth DTO 空壳删除实现方案

## Objective

删除已经无消费方的 `internal/dto/auth.go`。

## Non-goals

- 不改 auth 模块现有 request / response owner
- 不调整登录、注册、CAS 行为

## Inputs

- `code/backend/internal/dto/auth.go`
- `code/backend/internal/module/auth/**`

## Task slices

1. 确认无消费方
2. 删除 `internal/dto/auth.go`
3. 跑最小编译与相关测试

## Expected changes

- `code/backend/internal/dto/auth.go`

## Validation

- `go test ./internal/dto ./internal/module/auth/... -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- 是否确实没有残留 `dto.RegisterReq` / `dto.LoginResp` 等引用
- 删除后 auth 模块是否仍完全依赖模块内 owner
