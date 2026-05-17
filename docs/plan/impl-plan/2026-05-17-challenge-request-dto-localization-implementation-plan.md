# challenge request DTO 模块内化实现方案

## Objective

把 `challenge` 与 `challenge/awd` 的 HTTP request/query DTO 从全局 `internal/dto` 收回 `challenge/api/http`，保持路由、参数语义、校验规则和外部 JSON 契约不变。

## Non-goals

- 不迁移 challenge response DTO（仍暂由全局 dto 承载）
- 不改 command/query service 的输出类型
- 不改外部路径、状态码与 JSON 字段

## Inputs

- `code/backend/internal/module/challenge/api/http/*.go`
- `code/backend/internal/module/challenge/api/http/request_mapper*.go`
- `code/backend/internal/dto/{challenge.go,awd_challenge.go,tag.go,image.go,topology.go}`
- `docs/plan/impl-plan/2026-05-17-challenge-contest-instance-awd-dto-localization-next-batch-plan.md`

## Task Slices

1. 新增模块内 request/query DTO
   - 在 `challenge/api/http` 建立 `challenge_request_types.go`
   - 覆盖 challenge、awd challenge、tag、image、topology、flag 的入参和查询参数

2. 收口 request mapper 输入类型
   - `request_mapper.go` 的输入签名改为本地 DTO
   - 仅对仍消费全局 query 的链路保留 mapper 过渡转换（本地 query -> `dto.ChallengeQuery`）
   - 重新生成 `request_mapper_gen.go`

3. 替换 handler 绑定类型
   - 相关 handler `ShouldBindJSON/ShouldBindQuery` 改为本地 DTO
   - 保持 handler 行为与错误返回语义不变

## Expected Changes

- `code/backend/internal/module/challenge/api/http/challenge_request_types.go`
- `code/backend/internal/module/challenge/api/http/request_mapper.go`
- `code/backend/internal/module/challenge/api/http/request_mapper_gen.go`
- `code/backend/internal/module/challenge/api/http/{handler.go,tag_handler.go,image_handler.go,topology_handler.go,flag_handler.go,awd_challenge_handler.go}`
- `docs/architecture/backend/04-api-design.md`

## Validation

- `go generate ./internal/module/challenge/api/http`
- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`

## Review Focus

- challenge request/query owner 是否已回到 `challenge/api/http`
- handler 是否统一通过 mapper 转换，而非直接依赖全局 request DTO
- 迁移后外部接口绑定、参数校验和错误语义是否保持一致
