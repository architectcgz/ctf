# Reuse Decision

## Change type

service / port / infrastructure / composition / docs

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/challenge_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_query_repository.go`

## Decision

refactor_existing

## Reason

这次 debt 不是 challenge core command 缺少新的题目写能力，而是 `challenge/application/commands/challenge_service.go` 直接知道 `gorm.ErrRecordNotFound`，并且构造函数为了 import / export use case 里的事务访问，把 `gorm.DB` 一起留在了 core command surface。

最小正确方案不是改 raw `challengeinfra.Repository` 的全局 not-found 语义，也不是把 challenge import / package revision / image build 这些仍然合理依赖 `gorm` 的事务面一起塞进这刀里，而是沿用 phase5 已经稳定下来的 error adapter 模式：

- 在 `challenge/ports` 增加 core command use case 需要的模块语义：`ErrChallengeCommandChallengeNotFound`、`ErrChallengePublishCheckJobNotFound`
- 新增窄 `ChallengeCommandRepository` adapter，只把 core command path 上的 `FindByID` 与 publish-check lookup `gorm.ErrRecordNotFound` 收口成模块 sentinel
- `challenge_service.go` 只消费模块 sentinel；镜像与拓扑 lookup 继续复用已经存在的 `ImageQueryRepository` / `TopologyServiceRepository` adapter
- `challenge/runtime/module.go` 只给 core challenge command service 注入 adapted command/image/topology repo，package revision 仍保留 raw repo
- 为了真正删掉 `challenge_service.go -> gorm.io/gorm`，把仍需 `gorm.DB` 的 `ChallengeService` struct / `NewChallengeService` / `SelfCheckConfig` 放回已经允许依赖 `gorm` 的 `challenge_import_service.go`
- 因为 HTTP handler 复用同一个 `ChallengeService` 实例，`challenge_package_revision_service.go` 的 challenge/topology lookup 也要同时接受新的 challenge/topology sentinel；但 revision repo lookup 仍保留 raw repo 语义，不扩成新的 package revision adapter

这样可以把范围收在 core challenge command path，同时不改变 challenge import、package revision 和 image build 这些后续 slice 的 owner 与事务边界。

## Files to modify

- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_error_contract_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_self_check_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
- `code/backend/internal/module/challenge/infrastructure/challenge_command_repository_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/app/challenge_import_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `.harness/reuse-decisions/challenge-core-command-not-found-contract-phase5-slice35.md`
- `docs/plan/impl-plan/2026-05-13-challenge-core-command-not-found-contract-phase5-slice35-implementation-plan.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/design/backend-module-boundary-target.md`

## After implementation

- `challenge/application/commands/challenge_service.go -> gorm.io/gorm` 这条 application concrete allowlist 已删除
- 共享 `ChallengeService` 上的 core command 与 package export path 都已接受 challenge/topology sentinel；challenge import、package revision repo lookup 与 image build 仍保留在后续独立 slice 收口
