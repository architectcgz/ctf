# Reuse Decision

## Change type
test / architecture guardrail / backend system-test boundary

## Existing code searched
- `code/backend/internal/module/architecture_test.go`
- `code/backend/internal/app/architecture_rules_test.go`
- `code/backend/internal/app/backend_context_architecture_test.go`
- `code/backend/internal/app/*_test.go`
- `code/backend/tests/README.md`
- `code/backend/tests/system/http/`
- `scripts/check-backend-architecture.sh`
- `docs/architecture/README.md`
- `works/backend-test-architecture-rewrite-blueprint.md`

## Similar implementations found
- `code/backend/internal/module/architecture_test.go` 已经用源码扫描守住模块边界、allowlist 和 stale allowlist。
- `code/backend/internal/app/backend_context_architecture_test.go` 已经用源码扫描守住 `internal/app` 的 context / concrete import 边界。
- `scripts/check-backend-architecture.sh` 已经是后端架构守卫的统一入口，适合继续接入新的测试架构守卫，而不是再造一套独立脚本。
- `code/backend/tests/README.md` 和 `works/backend-test-architecture-rewrite-blueprint.md` 已经把目标测试分层说清楚，当前缺口在于缺少机械 guardrail 防止回流。

## Decision
refactor_existing

## Reason
这次不是再设计一份新的测试架构，而是把已经确定的测试 owner 分层机械化：

- 继续复用现有 Go 架构测试模式，在 `code/backend/tests/architecture/` 下新增源码扫描守卫。
- 继续复用 `scripts/check-backend-architecture.sh` 作为统一入口，把测试架构守卫纳入 `--full` 检查。
- 补 `code/backend/tests/README.md` 和 `docs/architecture/README.md`，让“系统测试 shell 应该薄、scenario package 不该持有 DB/env owner”这条规则有文档入口和脚本入口。

本轮不做：

- 不迁移新的系统测试场景。
- 不重写 `internal/app` 现有 full-router fixture。
- 不把 `tests/system/http` 一次性抽成更细的 testkit 层。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-guard.md`
- `docs/plan/impl-plan/2026-06-04-backend-test-architecture-guard-plan.md`
- `code/backend/tests/architecture/test_architecture_test.go`
- `code/backend/tests/README.md`
- `docs/architecture/README.md`
- `scripts/check-backend-architecture.sh`

## After implementation
- 后续如果再迁新的系统测试 owner，必须同时过这组测试架构守卫，而不是只改 README。
- 如果 `internal/app` 的兼容壳继续收薄，应同步收紧 oversize allowlist，而不是长期保留宽松例外。
