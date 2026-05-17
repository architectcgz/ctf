# Reuse Decision

## Change type

backend / command mapper dto localization

## Existing code searched

- `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/application/commands/awd_response_mappers.go`
- `code/backend/internal/module/contest/architecture_test.go`

## Similar implementations found

- 近期 contest command 输出契约已统一迁移到 `application/commands/contest_output.go`。
- AWD run/preview 输出已改为 command 本地类型，当前仅 mapper 内部仍引用全局 dto。

## Decision

refactor_existing

## Reason

为避免 `commands` 层继续依赖 `internal/dto`，本刀把 AWD 相关 base mapper 输出类型改为 command 本地类型，并通过架构守卫防止回流。

## Files to modify

- `.harness/reuse-decisions/contest-command-awd-response-mapper-dto-localization.md`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/contest/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/contest/application/commands/awd_response_mappers.go`
- `code/backend/internal/module/contest/architecture_test.go`
