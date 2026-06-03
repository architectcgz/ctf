# 2026-06-03 backend test architecture phase 4 plan

## Objective

- 把 `full_router_access` 三个系统测试的断言 owner 迁到 `code/backend/tests/system/http`。
- 让 `internal/app/full_router_access_integration_test.go` 只保留兼容入口和最小 glue code。
- 继续沿用 `practice_flow` 已验证的“先迁场景断言，后拆 env owner”节奏。

## Non-goals

- 本轮不抽取 `newFullRouterTestEnv` 到共享 testutil。
- 本轮不迁移 `full_router_admin_*`、`full_router_teacher_*`、`full_router_contest_*`、`full_router_awd_*`。
- 本轮不修既有 `TestFullRouter_TeacherAWDReviewExportStateMatrix` 的 PDF 断言失败。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_access_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/tests/system/http/practiceflow/practice_flow.go`
- `.harness/reuse-decisions/backend-test-architecture-phase4.md`

## Current problem

- `full_router_access_integration_test.go` 的场景断言还在 `internal/app`，虽然 env owner 本身在别的文件。
- 如果直接抽 full router env，会马上碰到大体量 seed、route materialize 和 report wait helper，scope 不适合这一刀。
- 但如果不继续迁，`internal/app` 仍然持有 full router 场景测试 owner，测试架构重做会停在半路。

## Working design

### Target structure

- 新增 `code/backend/tests/system/http/fullrouteraccess/`：
  - `VerifyAccessControlMatrix`
  - `VerifyAuthorizedSmokeMatrix`
  - `VerifyListInstancesMatchesContract`
- 新 package 只持有断言逻辑和小型 DTO / driver 定义。
- `internal/app/full_router_access_integration_test.go` 提供：
  - route/env helper 到新 package 的 glue code
  - 现有 `newFullRouterTestEnv`、`performFullRouterRequest`、`waitForReportStatus` 的桥接调用

### Boundary choice

- 断言 owner 先迁，env owner 暂不迁。
- 通过 callback/driver 复用现有 `internal/app` helper，而不是复制 full router fixture。
- 等 access 场景迁完，再判断是否值得把 full router env 抽成独立 testutil。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 4 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase4`

### Slice 2：迁移 full router access 场景断言 owner

- Goal：把 `full_router_access_integration_test.go` 的断言逻辑迁到 `tests/system/http/fullrouteraccess`。
- Touched files：
  - `code/backend/internal/app/full_router_access_integration_test.go`
  - `code/backend/tests/system/http/fullrouteraccess/*.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_AccessControlMatrix|TestFullRouter_AuthorizedSmokeMatrix|TestFullRouter_ListInstancesMatchesContract' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 practice/router 相关现有行为和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- full router access 场景测试 owner
- `tests/system/http` 下第二个真实专题目录

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - callback/driver 抽象不够清楚，导致断言包反而更难读
  - glue code 漏掉 `reports/class` 的异步等待逻辑
  - 新 package 仍隐式耦合 `internal/app` 的未导出语义，后续迁移价值不足

## Review fit check

- Owner 清晰：场景断言 owner 迁到 `tests/system/http/fullrouteraccess`。
- Reuse 清晰：继续复用现有 full router fixture，不在这一刀复制或重写 env。
- 结构收敛：这刀只解决 access 场景 owner，不假装一次解决 full router 全部测试基建。

## Rollback / recovery

- 纯测试断言 owner 调整，可直接回退到当前 `internal/app/full_router_access_integration_test.go` 版本。
