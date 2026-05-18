# Reuse Decision

## Change type

contract / api / app-test / dto cleanup

## Existing code searched

- `code/backend/internal/dto/{challenge.go,challenge_import.go,topology.go,awd_challenge.go}`
- `code/backend/internal/module/challenge/api/http/{challenge_request_types.go,challenge_response_types.go}`
- `code/backend/internal/app/{full_router_integration_test.go,full_router_state_matrix_integration_test.go}`

## Similar implementations found

- `challenge/api/http` 已完整持有 challenge / topology / import / awd challenge 的 request / response DTO
- 第一刀 `challenge-image-tag-dto-residual-cleanup` 已证明剩余全局 DTO 主要是 app 测试残留
- `challenge/contracts` 与 `challenge/api/http` 的 mapper / handler 路径已是当前 owner

## Decision

refactor_existing

## Reason

这批全局 DTO 文件里定义的类型，owner 已经全部落在 `challenge/api/http` 或 `challenge/contracts`。仓库内真实消费方只剩 app 集成测试还在沿用旧 `internal/dto` 命名，而 `awd_challenge_import.go` 只是继续拼接已经迁走的 challenge/import/awd 类型；因此最小正确方案是把测试改到 challenge 模块自己的 HTTP DTO，并直接删除整批残留文件，而不是再保留一个全局兼容层。

## Files to modify

- `.harness/reuse-decisions/challenge-http-dto-residual-cleanup.md`
- `docs/plan/impl-plan/2026-05-18-challenge-http-dto-residual-cleanup-implementation-plan.md`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/awd_challenge_import.go`
- `code/backend/internal/dto/challenge.go`
- `code/backend/internal/dto/challenge_import.go`
- `code/backend/internal/dto/topology.go`
- `code/backend/internal/dto/awd_challenge.go`

## After implementation

- app 集成测试不再依赖 challenge 相关全局 DTO
- `internal/dto` 仅剩共享 `common.go` 与测试文件
- challenge request / response DTO 只保留在模块 owner 下
