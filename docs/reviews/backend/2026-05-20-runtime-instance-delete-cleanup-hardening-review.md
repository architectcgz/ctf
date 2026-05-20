# Runtime Instance Delete Cleanup Hardening Review

- Review target: `ctf` repo，本地 `main` 工作区；review 范围为实例删除异步化与 `stopping` cleanup worker 相关 diff
- Files reviewed:
  - `code/backend/internal/module/instance/application/commands/instance_service.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/config/config_test.go`
  - `code/backend/configs/config.yaml`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - `code/backend/internal/module/runtime/infrastructure/repository_destroyed_at_test.go`
  - `code/backend/internal/module/runtime/application/instance_service_test.go`
  - `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`
  - `code/backend/internal/module/runtime/service_test.go`
  - `code/backend/internal/module/practice/ports/ports.go`
  - `code/backend/internal/module/practice/ports/instance_context_contract_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
  - `code/backend/internal/module/practice/application/commands/service_test.go`
  - `docs/architecture/backend/05-key-flows.md`
  - `.harness/reuse-decisions/runtime-instance-delete-cleanup-hardening.md`
  - `docs/plan/impl-plan/2026-05-20-runtime-instance-delete-cleanup-hardening-implementation-plan.md`
- Classification check: agree with pipeline，属于 non-trivial backend implementation + review gate
- Gate verdict: pass

## Findings

`no findings`

## Residual Risks

- 当前验证已经覆盖 repository 条件更新、provisioning 分支补偿清理、删除 worker 并发上限和异步删除主链路，但还没有一条真实并发的端到端测试，把“删除请求”和“provisioning worker”同时打到同一实例上验证最终态。
- `instance_stopping_cleanup` 目前是进程内并发控制。如果未来后端改成多副本同时运行这条 loop，正确性更多依赖 cleanup 幂等性和数据库状态条件，而不是分布式 claim；这是后续运行模型需要继续评估的点。

## Validation Evidence

- `cd code/backend && go test ./internal/config ./internal/module/runtime ./internal/module/runtime/application ./internal/module/runtime/application/commands ./internal/module/runtime/api/http ./internal/module/runtime/infrastructure ./internal/module/instance/... ./internal/module/practice/... -count=1`
- `cd code/backend && go test ./internal/module/runtime/infrastructure -run 'TestFailProvisioningDoesNotOverrideStoppingInstance|TestFinalizeStoppedRuntimeClearsRuntimeFieldsAndAllocations' -count=1`
- `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestMarkInstanceFailedSkipsFailedTransitionWhenInstanceLeavesCreating|TestProvisionInstanceCleansRuntimeWhenInstanceLeavesCreatingBeforePersist' -count=1`
- `bash scripts/check-consistency.sh`
