# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/submission_*.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter*.go`
- `code/backend/internal/module/contest/api/http/submission_handler.go`
- `code/backend/internal/module/contest/architecture_test.go`

## Similar implementations found

- `practice` 与 `assessment` 已把 command 输出 DTO 收回各自模块 output 类型
- `contest/api/http/submission_handler.go` 已以模块内 `contestcommands.SubmissionResp` 作为返回契约

## Decision

refactor_existing

## Reason

`contest` submission command 仍通过 `dto.SubmissionResp` 返回，导致 application 层输出 owner 漂到全局 `internal/dto`。本刀把输出收口到 `contest/application/commands/submission_output.go`，保持 HTTP 字段与行为不变，只迁移 owner。

## Files to modify

- `.harness/reuse-decisions/contest-submission-output-contract-localization.md`
- `code/backend/internal/module/contest/application/commands/submission_output.go`
- `code/backend/internal/module/contest/application/commands/submission_submit.go`
- `code/backend/internal/module/contest/application/commands/submission_incorrect_submit.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/architecture_test.go`
