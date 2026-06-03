# Reuse Decision

## Change type
backend test bugfix / runtime test environment hardening

## Existing code searched

- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/composition/runtime_test_engine.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`

## Similar implementations found

- `internal/app` 测试配置统一走 `newPracticeFlowTestConfig(t)`，当前 host port range 仍写死为 `30000-30100`。
- `runtime_test_engine.go` 在 test 环境会真实监听 `127.0.0.1:<hostPort>`，因此测试端口池必须避开宿主机已占用端口。
- 现有代码里还没有 `internal/app` 级别的测试端口分配 helper。

## Decision
refactor_existing

## Reason

当前最小正确修复不是放宽断言，也不是改业务 runtime 逻辑，而是修正测试环境假设：

- 这批 `internal/app` 集成测试在 test runtime 下会真实占用本机端口。
- 固定端口池 `30000-30100` 与宿主机已有服务冲突时，会把实例创建直接打成 500。
- 这个问题属于测试 harness 自己的环境假设，应该由共享测试 helper 统一分配安全端口范围，而不是在单个测试里绕开。

## Files to modify

- `.harness/reuse-decisions/backend-internal-app-test-runtime-port-range.md`
- `docs/plan/impl-plan/2026-06-03-backend-internal-app-test-runtime-port-range-plan.md`
- `code/backend/internal/app/test_schema_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

## After implementation

- `newPracticeFlowTestConfig(t)` 不再固定写死 `30000-30100`。
- `internal/app` 测试运行时端口范围会动态避开当前宿主机已占用端口，并在同进程测试间避免互相重叠。
- `TestFullRouter_AuthorizedSmokeMatrix` 与 `TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge` 至少不再因为端口冲突触发既有 500。
