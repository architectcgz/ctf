# contest request DTO 模块内化实现方案

## Objective

把 `contest` HTTP 请求/查询 DTO 从全局 `internal/dto` 收回 `contest/api/http`，保持所有竞赛与 AWD 路由、参数语义、校验规则不变。

## Non-goals

- 不迁移 contest response DTO（仍暂由全局 dto 承载）
- 不改 command/query service 输出类型
- 不改外部路径、状态码与 JSON 字段

## Inputs

- `code/backend/internal/module/contest/api/http/*.go`
- `code/backend/internal/module/contest/api/http/request_mapper*.go`
- `code/backend/internal/dto/{contest.go,team.go,contest_challenge.go,awd.go,contest_awd_service.go}`
- `docs/plan/impl-plan/2026-05-17-challenge-contest-instance-awd-dto-localization-next-batch-plan.md`

## Task Slices

1. 新增模块内 request/query DTO
   - 在 `contest/api/http` 建立 `contest_request_types.go`
   - 覆盖 contest/team/challenge/participation/awd/traffic/scoreboard 入参

2. 收口 request mapper 输入类型
   - `request_mapper.go` 的 input 映射签名改为本地 DTO
   - 重新生成 `request_mapper_gen.go`

3. 替换 handler 绑定类型
   - 各 handler `ShouldBindJSON/ShouldBindQuery` 改为本地 DTO
   - 保持 handler 行为与错误返回语义不变

## Expected Changes

- `code/backend/internal/module/contest/api/http/contest_request_types.go`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/*handler*.go`
- `docs/architecture/backend/04-api-design.md`

## Validation

- `go generate ./internal/module/contest/api/http`
- `go test ./internal/module/contest/... -count=1`

## Review Focus

- contest request DTO owner 是否已回到 `contest/api/http`
- 入参 tags、校验口径和 query 默认行为是否保持一致
- 迁移后是否没有把 request 类型漏回 `internal/dto`
