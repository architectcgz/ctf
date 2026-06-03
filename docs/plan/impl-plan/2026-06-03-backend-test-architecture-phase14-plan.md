# 2026-06-03 backend test architecture phase 14 plan

## Objective

- 把 `full_router_integration_test.go` 里的通用 full-router env / seed / request helper 抽到独立 test helper 文件。
- 让 `full_router_integration_test.go` 主要只保留 router/module 级测试与 access-specific helper。
- 继续沿用最小可审阅切片，不改 helper 语义。

## Non-goals

- 本轮不改 access route classify/materialize 相关 helper owner。
- 本轮不迁移任何新的 HTTP 场景断言 owner。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_*_test.go`
- `.harness/reuse-decisions/backend-test-architecture-phase14.md`

## Current problem

- `full_router_integration_test.go` 虽然已基本没有长场景测试，但仍承载大量被多文件复用的 env / seed / request helper。
- helper owner 和当前文件仅存的 module builder 测试没有同一职责。
- 继续把新 helper 堆回这个文件，会让“测试入口”和“测试基座”混杂。

## Working design

### Target structure

- 新增 `code/backend/internal/app/full_router_test_helpers_test.go`
  - `fullRouterTestEnv`
  - `newFullRouterTestEnv`
  - `openFullRouterTestDB`
  - `newFullRouterTestConfig`
  - `seedFullRouterData`
  - `seedRoles`
  - `createFullRouterUser`
  - `performFullRouterRequest`
- `full_router_integration_test.go` 保留：
  - `TestRouterBuildUsesCompositionModules`
  - access route classify / payload / materialize helper

### Boundary choice

- 只迁共享 helper owner，不改函数签名。
- 继续保留 `package app`，避免引入额外可见性调整。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 14 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase14`

### Slice 2：抽 shared full-router helper owner

- Goal：把共享 env / seed / request helper 挪到独立 test helper 文件。
- Touched files：
  - `code/backend/internal/app/full_router_integration_test.go`
  - `code/backend/internal/app/full_router_test_helpers_test.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`

### Slice 3：workflow gate

- Goal：确认整体 workflow 约束未被破坏。
- Validation：
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- `internal/app` 下新增 full router shared helper file
- `full_router_integration_test.go` 体量显著下降

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - helper 挪动后 import 清理不完整
  - 某些 full-router 测试遗漏依赖的 package import 或 symbol

## Review fit check

- Owner 清晰：full router 共享测试基座与单一测试入口分离。
- Reuse 清晰：不改 helper 签名，现有调用点无需跟着重写。
- 结构收敛：这刀只处理 shared helper owner，不碰 access/state/admin 场景边界。

## Rollback / recovery

- 纯测试 helper 重排，可直接回退到当前 `full_router_integration_test.go` 版本。
