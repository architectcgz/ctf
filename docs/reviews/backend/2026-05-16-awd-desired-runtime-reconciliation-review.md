# AWD Desired Runtime Reconciliation Review

- Review target: `ctf` repo，本地 `main` 工作区；review 范围为 `git diff -- code/backend/internal/config/config.go code/backend/configs/config.yaml code/backend/configs/config.prod.yaml code/backend/internal/module/practice/ports/ports.go code/backend/internal/module/practice/infrastructure/repository.go code/backend/internal/module/practice/application/commands/service.go code/backend/internal/module/practice/runtime/module.go code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go code/backend/internal/module/practice/application/commands/instance_start_service.go code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go code/backend/internal/app/composition/instance_module.go code/backend/internal/app/composition/practice_module.go docs/architecture/backend/03-container-architecture.md docs/architecture/backend/05-key-flows.md docs/plan/impl-plan/2026-05-16-awd-desired-runtime-reconciliation-implementation-plan.md .harness/reuse-decisions/awd-desired-runtime-reconciliation.md`
- Files reviewed: 同上
- Classification check: agree with pipeline，属于 non-trivial backend implementation + review gate
- Gate verdict: pass

## Findings

- 无新的 material finding。review 过程中发现过一个已修复回归：
  - [code/backend/internal/module/practice/application/commands/instance_start_service.go](/home/azhi/workspace/projects/ctf/code/backend/internal/module/practice/application/commands/instance_start_service.go:88)
  - 公共 helper 初版把“active instance 直接返回”也带进了用户主动重启路径，导致 `RestartContestAWDService` 在实例仍为 active 时不会真的重启。当前已通过 `NoopIfActive` 显式限定为 only desired reconcile 使用，并回归验证相关 restart tests。

## Material Findings

- 无

## Senior Implementation Assessment

- 当前实现把 owner 边界保持在正确位置：
  - `startup_runtime_recovery_service` 只做停机检测、暂停时间补偿和恢复顺序编排
  - `maintenance_service` 只做 active runtime repair
  - `practice` 负责 `team × visible service` 的期望态补齐
- 周期调和复用现有 `practice_instance_scheduler`，而不是再起一套并行控制面；这比把差集推导塞进 `maintenance_service` 或扩成“批量重启”入口更稳。

## Required Re-validation

- `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestReconcileDesiredAWDInstances|TestRunProvisioningLoop.*Desired|TestRestartContestAWDService(PreservesExistingDefenseWorkspaceRevision|RecreatesMissingDefenseWorkspaceContainer)' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/practice/application/commands -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/instance/application/commands -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/app/composition -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/config -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

## Residual Risk

- 当前 desired reconcile 对“永久坏配置”的 scope 仍会按 `container.scheduler.desired_reconcile_interval` 持续重试，没有 backoff / suppress 机制。这不影响“平台恢复后把应该活着的实例补齐”的主目标，但在题目配置长期错误时会带来重复 operation 记录和周期性噪声。
- 本次没有做整机停机后的端到端 docker 宿主恢复演练；当前证据来自 package tests、composition tests 和文档/一致性检查。

## Touched Known-Debt Status

- 本次 touched surface 未命中已记录且必须在本 slice 内强制收口的结构性 debt 记录。
