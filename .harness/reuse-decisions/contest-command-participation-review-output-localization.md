# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/participation_review_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter*.go`
- `code/backend/internal/module/contest/api/http/participation_handler.go`

## Similar implementations found

- 同模块 `submission`、`contest`、`team`、`challenge add` 已统一收口到 `contest/application/commands` output。
- `ReviewRegistration` 输出映射复用同一 mapper 模式。

## Decision

refactor_existing

## Reason

`ReviewRegistration` 仍返回全局 `dto.ContestRegistrationResp`。本刀把 review 输出收口到 `contest/application/commands`，保持 HTTP 字段与语义不变；`CreateAnnouncement` 保持现状，避免扩大本刀范围。

## Files to modify

- `.harness/reuse-decisions/contest-command-participation-review-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/participation_review_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/api/http/participation_handler.go`
- `code/backend/internal/module/contest/architecture_test.go`
