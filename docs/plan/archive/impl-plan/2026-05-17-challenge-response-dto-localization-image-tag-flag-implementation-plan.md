# challenge image/tag/flag response DTO 模块内化实现方案

## Objective

把 `challenge` 模块内 `image/tag/flag` 三组 HTTP 响应从全局 `internal/dto` 透传改为模块内 response DTO + response mapper，保持外部响应字段与错误语义不变。

## Non-goals

- 不迁移 challenge/topology/awd/challenge import 等其余 response DTO
- 不改 application command/query service 的输出类型
- 不改路由、状态码、鉴权和错误码

## Inputs

- `code/backend/internal/module/challenge/api/http/{tag_handler.go,image_handler.go,flag_handler.go}`
- `code/backend/internal/module/challenge/api/http/request_mapper*.go`
- `code/backend/internal/module/ops/api/http/notification_response_mapper*.go`
- `code/backend/internal/dto/{tag.go,image.go,challenge.go}`

## Task Slices

1. 新增模块内 response DTO
   - 在 `challenge/api/http` 新增 `challenge_response_types.go`
   - 覆盖 `TagResp`、`ImageResp`、`FlagResp` 和分页壳 `PageResult`

2. 新增 response mapper
   - 在 `challenge/api/http` 新增 `response_mapper.go`
   - 由 mapper 负责 `dto -> http dto` 转换
   - 对图片分页补最小手写壳转换，避免 handler 透传 `dto.PageResult`

3. 更新 handler 输出转换
   - `tag_handler`、`image_handler`、`flag_handler` 全部走 response mapper
   - 保持 `response.Success` 的 data 结构与字段命名不变

## Expected Changes

- `code/backend/internal/module/challenge/api/http/challenge_response_types.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
- `code/backend/internal/module/challenge/api/http/response_mapper_assign.go`
- `code/backend/internal/module/challenge/api/http/response_mapper_gen.go`
- `code/backend/internal/module/challenge/api/http/{tag_handler.go,image_handler.go,flag_handler.go}`

## Validation

- `go generate ./internal/module/challenge/api/http`
- `go test ./internal/module/challenge/api/http -count=1`
- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`

## Review Focus

- handler 是否不再直接返回全局 `dto.TagResp` / `dto.ImageResp` / `dto.FlagResp`
- 输出字段和分页 JSON key 是否保持现有契约
- mapper 是否遵循全局委托策略，未引入额外业务逻辑
