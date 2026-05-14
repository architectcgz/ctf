# Reuse Decision

## Change type

service / port / runtime / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`
- `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/repository.go` 的 `WithinTransaction`
- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_http_runtime_adapter.go`
- `code/backend/internal/module/challenge/application/commands/image_build_service.go` 在 slice42 引入的窄 tx store

## Decision

refactor_existing

## Reason

剩余 4 条 allowlist 都集中在 challenge 的三条 transaction-heavy command 路径上：

- `challenge_import_service.go`
- `awd_challenge_import_service.go`
- `challenge_package_revision_service.go`

继续沿用“application 里允许少量 `gorm` 特例”的中间态已经收益很低。更合适的做法是把事务 owner 从 application service 抽走，收口到 challenge 模块自己的 use-case-oriented tx runner / tx store：

- application 只保留流程控制、文件编排、errcode 映射和业务语义
- runtime 负责把 raw repo、image build 能力和 tx bridge 装起来
- infrastructure 只保留底层 GORM source，不反向依赖 application

不建议这次再造宽泛的全模块 tx store，也不建议继续扩大 raw repo 的全局错误语义。按 use case 切 `ChallengeImport` / `AWDChallengeImport` / `ChallengePackageExport` 三组 tx runner，owner 更明确，后续 review 也更容易做边界检查。

## Files to modify

- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_error_contract_test.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
- `code/backend/internal/module/challenge/application/commands/tx_runner_test.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/app/challenge_import_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/reviews/backend/2026-05-14-challenge-phase5-tx-runner-remaining-review.md`
- `docs/plan/impl-plan/2026-05-14-challenge-tx-runner-phase5-remaining-implementation-plan.md`

## Candidate new files

- `code/backend/internal/module/challenge/runtime/import_tx_bridge.go`
- `code/backend/internal/module/challenge/runtime/awd_import_tx_bridge.go`
- `code/backend/internal/module/challenge/runtime/package_export_tx_bridge.go`

## After implementation

- `challenge/application/commands/{challenge_import_service.go,awd_challenge_import_service.go,challenge_package_revision_service.go}` 的 `gorm` / `clause` allowlist 应可一次收完
- challenge 模块 phase5 剩余 concrete leak 将从“application 事务面”切换到更清晰的 tx runner / tx store 契约
- 若后续再处理长事务内的文件副作用或 registry verify，应该沿这次 tx bridge 继续拆 staging / finalize，而不是重新把业务塞回 application
