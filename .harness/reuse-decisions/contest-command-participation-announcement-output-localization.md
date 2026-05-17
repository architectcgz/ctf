# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/participation_announcement_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter*.go`
- `code/backend/internal/module/contest/api/http/participation_handler.go`

## Similar implementations found

- 同模块 `submission`、`contest`、`team`、`challenge`、`participation review` 已收口到模块内 output。
- `CreateAnnouncement` 与现有 command 输出映射模式一致。

## Decision

refactor_existing

## Reason

`CreateAnnouncement` 仍返回全局 `dto.ContestAnnouncementResp`。本刀把 announcement 输出收口到 `contest/application/commands`，保持外部 JSON 字段和事件 payload 语义不变。

## Files to modify

- `.harness/reuse-decisions/contest-command-participation-announcement-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/participation_announcement_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/api/http/participation_handler.go`
- `code/backend/internal/module/contest/architecture_test.go`
