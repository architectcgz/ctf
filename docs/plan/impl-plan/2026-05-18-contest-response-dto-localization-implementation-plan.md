# contest response DTO 模块内化实现方案

## Objective

把 `contest/api/http` 仍通过全局 `internal/dto` 暴露的 response DTO 收回 `contest` 模块内部，明确：

- request mapper 只负责 `api/http -> application` 入参映射
- response mapper 只负责 `application/domain -> api/http` 响应映射
- `contest/api/http` 不再直接 import `internal/dto`

## Non-goals

- 不改外部 HTTP 路由、状态码、JSON 字段和分页语义
- 不改 `contest/application/{commands,queries}` 已经收口过的 output owner
- 不删除 `internal/dto` 中仍被其他模块使用的存量类型

## Inputs

- `docs/plan/impl-plan/2026-05-17-contest-request-dto-localization-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-17-challenge-contest-instance-awd-dto-localization-next-batch-plan.md`
- `code/backend/internal/module/contest/api/http/*.go`
- `code/backend/internal/dto/{team.go,contest.go,contest_challenge.go,awd.go,contest_awd_workspace.go,common.go}`

## Task Slices

1. 新增 `contest/api/http` 本地 response DTO
   - 覆盖 team / participation / challenge / scoreboard / AWD workspace / readiness / traffic / summary 等 response 类型
   - 保持 JSON field 与 pointer 语义不变

2. 拆分 request / response mapper owner
   - `request_mapper.go` 只保留 request -> input 映射
   - 新增 `response_mapper.go` 承接 query/domain -> http response 映射
   - 重新生成 `request_mapper_gen.go` 与 `response_mapper_gen.go`

3. 切换 handler / support 到 response mapper
   - 所有 response 路径改用 `contestResponseMapper`
   - contest list / AWD service support / audit payload 一并切到新 owner

4. 收口 app 集成测试的 contest response decode owner
   - `internal/app` 命中 contest 路由的解码类型改为 `contest/api/http` 本地 response DTO
   - 不再让 app 测试继续依赖全局 `dto` 作为 contest HTTP response owner

5. 补架构守卫
   - `contest/api/http/*.go` 整体禁止 import `internal/dto`
   - 维持 `contest` 模块当前 request / output owner 不回流

## Expected Changes

- `code/backend/internal/module/contest/api/http/contest_response_types.go`
- `code/backend/internal/module/contest/api/http/response_mapper.go`
- `code/backend/internal/module/contest/api/http/response_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/*handler*.go`
- `code/backend/internal/module/contest/api/http/request_mapper_*_support.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/module/contest/architecture_test.go`

## Compatibility

- 外部接口字段、分页包装、AWD 目录/流量/态势响应结构保持不变
- `contestcmd.ContestResp`、`contestcmd.ContestAWDServiceResp` 等已在模块内的 output 不额外迁移

## Validation

- `cd code/backend && go generate ./internal/module/contest/api/http`
- `cd code/backend && go test ./internal/module/contest/api/http -count=1`
- `cd code/backend && go test ./internal/module/contest/... -count=1`
- `cd code/backend && go test ./internal/app -run 'TestFullRouter_ContestRegistrationParticipationAndTeamStateMatrix|TestFullRouter_AWDTrafficSummaryAndEventsStateMatrix|TestFullRouter_ContestChallengeAndScoreboardStateMatrix|TestFullRouter_AdminContestListUsesRegisteringCount' -count=1`
- `cd code/backend && go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`

## Review Focus

- `contest/api/http` 是否已完全脱离 `internal/dto`
- request mapper / response mapper owner 是否清晰，不再混放
- 本地 response DTO 是否保持既有 JSON shape 和 optional/pointer 语义

## Rollback

- 若 response DTO 本地化引发兼容性回归，可先回退 `contest/api/http` 本地 response types 与 mapper 拆分，恢复使用全局 `internal/dto`，再按更小切片重做
