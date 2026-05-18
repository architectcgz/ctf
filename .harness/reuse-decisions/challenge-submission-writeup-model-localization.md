# Reuse Decision

## Change type
entity / ports / repository / mapper / app-test / model localization

## Existing code searched
- `code/backend/internal/model/submission_writeup.go`
- `code/backend/internal/module/challenge/...`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `.harness/reuse-decisions/*challenge*writeup*`

## Similar implementations found
- `code/backend/internal/module/challenge/entity/package_revision.go`
- `code/backend/internal/module/challenge/entity/hint.go`
- `code/backend/internal/module/contest/entity/status_transition.go`

## Decision
refactor_existing

## Reason
`SubmissionWriteup` 是 challenge 模块题解投稿、教师审核和社区题解查询链路的持久化实体，写路径 owner 明确在 `challenge`。最小正确方案是把实体和其状态常量收回 `internal/module/challenge/entity`，保持表结构、审核语义和对外响应字段不变。

非目标：本刀不处理 `Submission`、`ChallengeWriteup`、`ContestRegistration`。

## Files to modify
- `code/backend/internal/model/submission_writeup.go`
- `code/backend/internal/module/challenge/entity/submission_writeup.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/ports/challenge_writeup_context_contract_test.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_repository.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_service_repository_test.go`
- `code/backend/internal/module/challenge/application/commands/writeup_service.go`
- `code/backend/internal/module/challenge/application/commands/writeup_service_context_test.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service.go`
- `code/backend/internal/module/challenge/application/queries/writeup_service_test.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/testsupport/test_helper.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## After implementation
- 删除 `internal/model/submission_writeup.go`
- 同步更新受影响的 goverter 生成代码
