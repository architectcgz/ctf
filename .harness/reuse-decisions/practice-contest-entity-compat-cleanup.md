# Reuse Decision

## Change type
ports / repository / application / test / entity owner compatibility cleanup

## Existing code searched
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/application/commands`
- `code/backend/internal/module/practice/infrastructure`
- `code/backend/internal/module/practice/testsupport/test_helper.go`
- `code/backend/internal/model/contest.go`
- `code/backend/internal/model/contest_registration.go`
- `code/backend/internal/model/team.go`
- `code/backend/internal/model/submission.go`
- `code/backend/internal/model/contest_challenge.go`
- `code/backend/internal/model/contest_awd_service.go`
- `code/backend/internal/model/contest_awd_service_snapshot.go`
- `code/backend/internal/module/contest/entity`

## Similar implementations found
- `code/backend/internal/module/contest`
- `code/backend/internal/module/runtime`
- `code/backend/internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`

## Decision
refactor_existing

## Reason
上一刀已经把 contest 主干持久化实体 owner 收回 `internal/module/contest/entity`，但 `practice` 仍通过
`internal/model` 的兼容 alias 读取这些实体。继续保留这层兼容会让 `practice` 看起来仍依赖共享层，
也会拖慢后面删除 `internal/model` 里 contest 兼容入口的节奏。

这刀只收口 `practice` 中 owner 已明确属于 contest 的实体引用：

- `Contest`
- `ContestRegistration`
- `ContestChallenge`
- `ContestAWDService`
- `ContestAWDServiceSnapshot`
- `Team`
- `TeamMember`
- `Submission`

仍然留在 `model` 的共享 owner 类型，例如 `Challenge`、`ChallengeTopology`、`Instance`、
`AWDDefenseWorkspace`、`AWDScopeControl`、`AWDServiceOperation`、`User` 等，本刀不迁移。

## Files to modify
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
- `code/backend/internal/module/practice/application/commands/awd_runtime_rules.go`
- `code/backend/internal/module/practice/application/commands/awd_scope_control_commands.go`
- `code/backend/internal/module/practice/application/commands/awd_scope_control_commands_test.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_service_native_test.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- `code/backend/internal/module/practice/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/practice/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- `code/backend/internal/module/practice/application/commands/score_service_test.go`
- `code/backend/internal/module/practice/application/commands/service_lifecycle_test.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`
- `code/backend/internal/module/practice/application/commands/submission_history_service.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository_test.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository_test.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `code/backend/internal/module/practice/infrastructure/score_repository.go`
- `code/backend/internal/module/practice/infrastructure/solved_submission_repository.go`
- `code/backend/internal/module/practice/infrastructure/solved_submission_repository_test.go`
- `code/backend/internal/module/practice/testsupport/test_helper.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## After implementation
- `practice` 内部不再通过 `internal/model` 引用 contest-owned 实体
- `practice` 仍可继续使用共享 owner 的 `model.*`
- 为下一刀继续清理 `runtime / assessment / app` 中的兼容 alias 留出直接路径
