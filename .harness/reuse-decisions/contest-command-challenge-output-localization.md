# Reuse Decision

## Change type

backend / command output dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/challenge_add_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter*.go`
- `code/backend/internal/module/contest/application/commands/response_mappers.go`
- `code/backend/internal/module/contest/api/http/challenge_handler.go`

## Similar implementations found

- 同模块 `submission`、`contest`、`team` 命令输出都已收口到 `contest/application/commands`。
- `challenge_add` 的输出映射结构与上述模式一致。

## Decision

refactor_existing

## Reason

`AddChallengeToContest` 仍返回全局 `dto.ContestChallengeResp`，导致 command 输出 owner 依赖全局 DTO。把输出收口到模块内，保持字段与行为不变，统一 command 输出边界。

## Files to modify

- `.harness/reuse-decisions/contest-command-challenge-output-localization.md`
- `code/backend/internal/module/contest/application/commands/contest_output.go`
- `code/backend/internal/module/contest/application/commands/challenge_add_commands.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/application/commands/response_mappers.go`
- `code/backend/internal/module/contest/api/http/challenge_handler.go`
- `code/backend/internal/module/contest/architecture_test.go`
