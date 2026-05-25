# Reuse Decision

## Change type

api / request mapper / dto localization

## Existing code searched

- `code/backend/internal/module/contest/api/http/*.go`
- `code/backend/internal/module/contest/api/http/request_mapper*.go`
- `code/backend/internal/dto/{contest.go,team.go,contest_challenge.go,awd.go,contest_awd_service.go}`
- `docs/plan/archive/impl-plan/2026-05-17-challenge-contest-instance-awd-dto-localization-next-batch-plan.md`

## Similar implementations found

- `identity/api/http` 的 admin user request/response DTO 已收回模块边界
- `assessment/api/http`、`practice/api/http` 已采用“模块内 request type + mapper”模式

## Decision

refactor_existing

## Reason

`contest` HTTP 入参长期挂在全局 `internal/dto`，导致 request owner 不清晰。先把 request/query 类型收回 `contest/api/http`，保持外部契约不变，再继续后续 response/output 切片。

## Files to modify

- `.harness/reuse-decisions/contest-request-dto-localization.md`
- `docs/plan/archive/impl-plan/2026-05-17-contest-request-dto-localization-implementation-plan.md`
- `code/backend/internal/module/contest/api/http/contest_request_types.go`
- `code/backend/internal/module/contest/api/http/request_mapper.go`
- `code/backend/internal/module/contest/api/http/request_mapper_gen.go`
- `code/backend/internal/module/contest/api/http/*handler*.go`
- `docs/architecture/backend/04-api-design.md`
