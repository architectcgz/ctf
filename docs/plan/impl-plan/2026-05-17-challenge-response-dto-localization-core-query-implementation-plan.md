# challenge core query response DTO 模块内化实现方案

## Objective

把 `challenge/api/http/handler.go` 中主查询面与导入提交结果的输出从全局 `internal/dto` 透传改为模块内 response DTO + response mapper，保持外部字段和行为不变。

## Non-goals

- 不迁移 `preview import`、`self-check`、`publish-check`、`package export` 响应
- 不改 application command/query service 输出类型
- 不改路由、鉴权、状态码与错误码

## Inputs

- `code/backend/internal/module/challenge/api/http/handler.go`
- `code/backend/internal/module/challenge/api/http/response_mapper*.go`
- `code/backend/internal/dto/{challenge.go,challenge_import.go,common.go}`

## Task Slices

1. 扩展模块内 response DTO
   - 增加 `ChallengeResp`、`ChallengeListItem`、`ChallengeDetailResp`、`ChallengeImportCommitResp`
   - 增加对应 hint 子类型

2. 扩展 response mapper
   - 增加 challenge 详情/列表/分页映射方法
   - 增加 `ChallengeImportCommitResp` 薄包装函数

3. 更新 handler 输出转换
   - `CreateChallenge`、`GetChallenge`、`ListChallenges`
   - `ListPublishedChallenges`、`GetPublishedChallenge`
   - `CommitChallengeImport`

## Expected Changes

- `code/backend/internal/module/challenge/api/http/challenge_response_types.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
- `code/backend/internal/module/challenge/api/http/response_mapper_gen.go`
- `code/backend/internal/module/challenge/api/http/handler.go`

## Validation

- `go generate ./internal/module/challenge/api/http`
- `go test ./internal/module/challenge/api/http -count=1`
- `go test ./internal/module/challenge/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`

## Review Focus

- handler 是否不再直接透传主 challenge 查询面的全局 dto
- list / page / detail / commit 的 JSON 字段与分页 key 是否保持不变
- mapper 是否仍然是机械映射而非业务逻辑分支
