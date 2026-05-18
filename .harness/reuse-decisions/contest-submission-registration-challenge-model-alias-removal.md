# Reuse Decision

## Change type
cleanup / compatibility shim removal / cross-module entity owner convergence / owner persistence entity relocation

## Existing code searched
- `code/backend/internal/model/submission.go`
- `code/backend/internal/model/contest_registration.go`
- `code/backend/internal/model/contest_challenge.go`
- `code/backend/internal/model/awd_challenge.go`
- `code/backend/internal/module/challenge/entity`
- `code/backend/internal/module/challenge/contracts`
- `code/backend/internal/module/contest/entity`
- `code/backend/internal/module/assessment`
- `code/backend/internal/module/teaching_query`
- `code/backend/internal/app`
- `docs/plan/impl-plan/2026-05-18-contest-core-persistence-entity-owner-localization-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-18-practice-contest-entity-compat-cleanup-implementation-plan.md`

## Similar implementations found
- `code/backend/internal/module/practice`
  - 已经直接依赖 `contest/entity`，说明外部模块可以消费 contest owner 实体，而不需要继续经由 `internal/model`。
- `code/backend/internal/module/contest`
  - owner 模块内部已经完成 `Submission / ContestRegistration / ContestChallenge` 的实体本地化。
- `code/backend/internal/model/port_allocation.go`
  - 已有同类兼容 alias 删除切片，可复用“先清调用，再删 shim”的路径。

## Decision
refactor_existing

## Execution notes
- 第 1 刀已提交 `d6ddf3cf`：先暴露 `contest/contracts` 稳定入口并替换外部调用面。
- 第 2 刀继续删除 `internal/model` 中剩余的三份兼容 shim，不再保留过渡别名。
- 同步把 `AWDChallenge` 从共享层移入 `challenge` owner，由 `challenge/contracts` 暴露跨模块读取入口。

## Reason
`Submission`、`ContestRegistration`、`ContestChallenge` 的真实 owner 已经明确在
`internal/module/contest/entity`。当前 `internal/model` 中这三个文件只剩兼容 alias / 常量转发，
继续保留会让 `assessment`、`teaching_query` 和 `app` 测试看起来仍依赖共享层，也会拖慢后续
`internal/model` contest 残留的整体删除。

最小正确方案不是继续保留全局 `internal/model`，也不是让外部模块直接引用 `contest/entity`。
本刀改为由 `contest/contracts` 暴露稳定契约入口，把仍使用 alias 的调用面切到 contract，再删除三份兼容入口。
仍然属于共享 owner 的 `User`、`Challenge`、`Contest`、`Team` 等类型不在本刀处理范围。

## Files to modify
- `code/backend/internal/model/submission.go`
- `code/backend/internal/model/contest_registration.go`
- `code/backend/internal/model/contest_challenge.go`
- `code/backend/internal/model/awd_challenge.go`
- `code/backend/internal/module/challenge/entity/awd_challenge.go`
- `code/backend/internal/module/challenge/contracts/persistence.go`
- `code/backend/internal/module/contest/contracts/persistence.go`
- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/backend/internal/module/assessment/infrastructure/repository_test.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository_test.go`
- `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/assessment/application/queries/teacher_awd_review_service_test.go`
- `code/backend/internal/module/challenge/domain/awd_package_parser.go`
- `code/backend/internal/module/challenge/domain/mappers.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter_gen.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_command_facade_test.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_service_test.go`
- `code/backend/internal/module/challenge/application/commands/tx_runner_test.go`
- `code/backend/internal/module/challenge/application/commands/writeup_submission_service_test.go`
- `code/backend/internal/module/challenge/application/queries/awd_challenge_service_context_test.go`
- `code/backend/internal/module/challenge/application/queries/awd_challenge_service_test.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository_test.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/challenge/infrastructure/repository_test.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/ports/awd_challenge_command_context_contract_test.go`
- `code/backend/internal/module/challenge/ports/awd_challenge_query_context_contract_test.go`
- `code/backend/internal/module/challenge/runtime/awd_import_tx_bridge.go`
- `code/backend/internal/module/challenge/testsupport/test_helper.go`
- `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
- `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support_contract_test.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_native_test.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
- `code/backend/internal/module/contest/application/commands/contest_challenge_error_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository_test.go`
- `code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter.go`
- `code/backend/internal/module/contest/infrastructure/contest_awd_challenge_lookup_adapter_test.go`
- `code/backend/internal/module/contest/testsupport/db.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository_test.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/infrastructure/proxy_traffic_recorder_test.go`
- `code/backend/internal/module/runtime/application/instance_service_test.go`
- `code/backend/internal/module/runtime/service_test.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/backend/internal/middleware/awd_readiness_audit.go`
- `code/backend/internal/middleware/audit_test.go`
- `code/backend/internal/middleware/awd_readiness_audit_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/plan/impl-plan/2026-05-18-contest-submission-registration-challenge-model-alias-removal-implementation-plan.md`

## After implementation
- `internal/model` 不再保留 `Submission`、`ContestRegistration`、`ContestChallenge` 兼容 alias
- 外部模块统一通过 `contest/contracts` 读取 `Submission`、`ContestRegistration`、`ContestChallenge` 及相关状态常量
- `teaching_query -> contest`、`challenge -> contest` 依赖被显式纳入受控 allowlist
