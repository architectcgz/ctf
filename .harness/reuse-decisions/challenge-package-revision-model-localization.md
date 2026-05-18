# Reuse Decision

## Change type
entity / ports / repository / runtime / mapper / app-test / model localization

## Existing code searched
- `code/backend/internal/model/challenge_package_revision.go`
- `code/backend/internal/module/challenge/...`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `.harness/reuse-decisions/*challenge*package*revision*`

## Similar implementations found
- `code/backend/internal/module/challenge/entity/hint.go`
- `code/backend/internal/module/contest/entity/status_transition.go`
- `code/backend/internal/module/ops/entity/audit_log.go`

## Decision
refactor_existing

## Reason
`ChallengePackageRevision` 是 challenge 题包导入、导出和拓扑版本链路的持久化实体，owner 明确在 `challenge` 模块。最小正确方案是把实体和其 source type 常量收回 `internal/module/challenge/entity`，保持表结构、版本行为和导入导出语义不变。

非目标：本刀不处理 `Challenge`、`ChallengeTopology`、`EnvironmentTemplate`。

## Files to modify
- `code/backend/internal/model/challenge_package_revision.go`
- `code/backend/internal/module/challenge/entity/package_revision.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/ports/challenge_topology_context_contract_test.go`
- `code/backend/internal/module/challenge/runtime/import_tx_bridge.go`
- `code/backend/internal/module/challenge/runtime/package_export_tx_bridge.go`
- `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`
- `code/backend/internal/module/challenge/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/application/commands/tx_runner_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
- `code/backend/internal/module/challenge/application/queries/topology_service_test.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository.go`
- `code/backend/internal/module/challenge/infrastructure/topology_service_repository_test.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_repository.go`
- `code/backend/internal/module/challenge/testsupport/test_helper.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

## After implementation
- 删除 `internal/model/challenge_package_revision.go`
- 同步更新受影响的 goverter 生成代码
