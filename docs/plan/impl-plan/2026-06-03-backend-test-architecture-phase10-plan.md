# 2026-06-03 backend test architecture phase 10 plan

## Objective

- 把 `TestFullRouter_ReportPreviewAndDownloadStateMatrix` 的 HTTP 场景断言 owner 迁到 `code/backend/tests/system/http`。
- 让 `internal/app/full_router_state_matrix_integration_test.go` 只保留 module builder 测试、共享 helper 和最小 glue code。
- 继续沿用前几轮已验证的“先迁断言 owner，后拆 helper owner”节奏。

## Non-goals

- 本轮不抽取 `createReportRecord`、`waitForReportStatus`、ZIP/PDF helper 到共享 testutil。
- 本轮不迁移 `TestTeacherRoutesAreServedByTeachingQuery`。
- 本轮不迁移 `TestStudentPracticeReadRoutesAreServedByPracticeModule`。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouterconteststate/contest_state.go`
- `.harness/reuse-decisions/backend-test-architecture-phase10.md`

## Current problem

- `full_router_state_matrix_integration_test.go` 仍把 report preview/download 的长 HTTP 场景断言放在 `internal/app`。
- 同文件还混着 module builder 测试和大量共享 helper owner，不适合一刀全抽。
- 如果把 helper owner 和场景 owner 混做，会让这一轮 scope 过宽，review 成本升高。

## Working design

### Target structure

- 新增 `code/backend/tests/system/http/fullrouterreportstate/`：
  - `VerifyReportPreviewAndDownloadStateMatrix`
- 新 package 只持有 HTTP 请求序列、断言和小型 driver 定义。
- `internal/app/full_router_state_matrix_integration_test.go` 继续负责：
  - 构造 `newFullRouterTestEnv`
  - 创建 processing/failed report records
  - 提供 `waitForReportStatus` callback
  - 保留 module builder 测试和共享 helper

### Boundary choice

- 断言 owner 先迁，helper owner 暂不迁。
- 通过 callback/driver 复用现有 `internal/app` helper，而不是复制报表 seed/wait helper。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 10 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase10`

### Slice 2：迁移 report preview/download 场景断言 owner

- Goal：把 `TestFullRouter_ReportPreviewAndDownloadStateMatrix` 迁到 `tests/system/http/fullrouterreportstate`。
- Touched files：
  - `code/backend/internal/app/full_router_state_matrix_integration_test.go`
  - `code/backend/tests/system/http/fullrouterreportstate/*.go`
  - `code/backend/tests/README.md`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_ReportPreviewAndDownloadStateMatrix' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 practice/router 相关现有行为和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- report preview/download 场景测试 owner
- `tests/system/http` 下第八个专题目录

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - processing/failed/ready 三种 report 状态断言迁出后出现语义漂移
  - class/contest/review archive 的权限断言遗漏
  - helper owner 暂留 `internal/app` 导致文件体量仍然偏大，但这是本轮有意保留的边界

## Review fit check

- Owner 清晰：report preview/download 的 HTTP 场景断言迁到 `tests/system/http/fullrouterreportstate`。
- Reuse 清晰：继续复用现有 full router fixture 和 report helper，不在这一刀复制 env 或 wait helper。
- 结构收敛：这刀只解决 report state owner，不把 helper 抽取混进来。

## Rollback / recovery

- 纯测试断言 owner 调整，可直接回退到当前 `internal/app/full_router_state_matrix_integration_test.go` 版本。
