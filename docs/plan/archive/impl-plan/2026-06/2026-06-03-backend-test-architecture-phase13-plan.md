# 2026-06-03 backend test architecture phase 13 plan

## Objective

- 把 `TestTeacherRoutesAreServedByTeachingQuery` 和 `TestStudentPracticeReadRoutesAreServedByPracticeModule` 从 `full_router_state_matrix_integration_test.go` 拆到独立文件。
- 让 `full_router_state_matrix_integration_test.go` 只保留 report state 场景 glue 和共享 helper。
- 继续沿用“小切片、最小可审阅 diff”的节奏。

## Non-goals

- 本轮不迁移 `waitForReportStatus`、`performFullRouterMultipartRequest`、`receiveFullRouterWSMessageByType` 等 helper owner。
- 本轮不改 `TestFullRouter_ReportPreviewAndDownloadStateMatrix` 的场景边界。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `.harness/reuse-decisions/backend-test-architecture-phase13.md`

## Current problem

- `full_router_state_matrix_integration_test.go` 里混着 module wiring 测试、report state 场景 glue 和大量共享 helper。
- wiring 测试和 report state 场景没有同一个 owner，继续放一起会让文件职责不清。
- 如果现在硬抽 helper owner，会把本轮 scope 扩大到不必要的程度。

## Working design

### Target structure

- 新增 `code/backend/internal/app/full_router_module_wiring_test.go`
  - `TestTeacherRoutesAreServedByTeachingQuery`
  - `TestStudentPracticeReadRoutesAreServedByPracticeModule`
- `full_router_state_matrix_integration_test.go` 删除上述两个测试，仅保留：
  - `TestFullRouter_ReportPreviewAndDownloadStateMatrix`
  - report / contest / instance 等共享 test helper

### Boundary choice

- 只做文件级 owner 拆分。
- 继续保留 `package app`，不引入新的测试 package 或目录。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 13 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase13`

### Slice 2：拆分 module wiring 测试文件

- Goal：把两个 module wiring 测试移到新文件。
- Touched files：
  - `code/backend/internal/app/full_router_state_matrix_integration_test.go`
  - `code/backend/internal/app/full_router_module_wiring_test.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestTeacherRoutesAreServedByTeachingQuery|TestStudentPracticeReadRoutesAreServedByPracticeModule' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 router / practice 相关回归和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- `internal/app` 下一个新的 wiring 测试文件
- `full_router_state_matrix_integration_test.go` 职责收敛

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - 拆分后 import 或 builder override 清理不完整
  - `TestPracticeFlow_|TestRouter` 回归时出现命名或依赖遗漏

## Review fit check

- Owner 清晰：module wiring 测试与 report state/helper 不再混在同一个文件。
- Reuse 清晰：不新造 helper，不改变现有 env / callback 结构。
- 结构收敛：本轮只做文件职责拆分，不碰共享 helper owner。

## Rollback / recovery

- 纯测试文件重排，可直接回退到当前 `full_router_state_matrix_integration_test.go` 版本。
