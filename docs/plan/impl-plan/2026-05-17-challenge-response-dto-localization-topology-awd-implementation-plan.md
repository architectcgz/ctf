# challenge topology / awd response DTO 模块内化实现方案

## Objective

把 `challenge` 模块内 `topology` 与 `awd_challenge` 两组 HTTP 响应从全局 `internal/dto` 透传改为模块内 response DTO + response mapper，保持外部响应字段和行为不变。

## Non-goals

- 不迁移主 challenge（authoring/query/import/self-check）响应 DTO
- 不改 application command/query service 输出类型
- 不改路由、状态码、鉴权与错误码

## Inputs

- `code/backend/internal/module/challenge/api/http/{topology_handler.go,awd_challenge_handler.go}`
- `code/backend/internal/module/challenge/api/http/response_mapper*.go`
- `code/backend/internal/dto/{topology.go,awd_challenge.go,awd_challenge_import.go,challenge_import.go}`
- `docs/plan/impl-plan/2026-05-17-challenge-response-dto-localization-image-tag-flag-implementation-plan.md`

## Task Slices

1. 扩展模块内 response DTO
   - 在 `challenge_response_types.go` 增加 topology 与 awd challenge 响应类型
   - 覆盖 topology/template、awd challenge list/detail/import preview/import commit

2. 扩展 response mapper
   - 在 `response_mapper.go` 增加 `dto -> http dto` 转换方法
   - 增加 awd page/commit 响应的薄包装函数

3. 更新 handler 输出转换
   - `topology_handler` 与 `awd_challenge_handler` 的所有 `response.Success` 输出统一走 mapper
   - 保持外部字段名与分页结构不变

## Expected Changes

- `code/backend/internal/module/challenge/api/http/challenge_response_types.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
- `code/backend/internal/module/challenge/api/http/response_mapper_gen.go`
- `code/backend/internal/module/challenge/api/http/{topology_handler.go,awd_challenge_handler.go}`
- `docs/architecture/backend/04-api-design.md`

## Validation

- `go generate ./internal/module/challenge/api/http`
- `go test ./internal/module/challenge/api/http -count=1`
- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`

## Review Focus

- handler 是否不再直接透传 topology/awd 相关全局 dto
- awd list/import 与 topology/template 的 JSON 字段是否保持一致
- mapper 扩展是否保持机械映射，不混入业务逻辑
