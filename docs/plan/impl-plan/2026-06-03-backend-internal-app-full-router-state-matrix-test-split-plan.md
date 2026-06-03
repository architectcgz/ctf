# 2026-06-03 backend internal app full router state matrix test split plan

## Objective

- 把 `code/backend/internal/app/full_router_state_matrix_integration_test.go` 中混杂的状态矩阵测试按领域拆成多个 `_test.go` 文件。
- 降低单文件体积和 owner 混杂程度，让后续继续维护 `internal/app` 集成测试时更容易定位。
- 保持测试行为、fixture、helper 和包内共享方式不变。

## Non-goals

- 本轮不改业务路由、handler、service 或测试断言语义。
- 本轮不重写 `newFullRouterTestEnv`、共享 schema helper 或 report helper。
- 本轮不顺手拆 `full_router_integration_test.go`、`practice_flow_integration_test.go`、`router_test.go`。

## Inputs

- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/router_test.go`
- `.harness/reuse-decisions/backend-internal-app-full-router-state-matrix-test-split.md`

## Current problem

- `full_router_state_matrix_integration_test.go` 当前接近 3000 行。
- 顶层测试同时覆盖 contest、teacher、admin、awd 多类责任，review 和定位回归时上下文噪声很大。
- helper 已经相对稳定，主要问题是顶层 `Test...` owner 没有收口到独立文件。

## Working design

### Target structure

- 保留原文件作为 shared helper owner：
  - 两个轻量 wiring 测试
  - `TestFullRouter_ReportPreviewAndDownloadStateMatrix`
  - 底部所有 `assert/decode/create/...` helper
- 新增 4 个按领域归类的测试文件：
  - `full_router_contest_state_matrix_test.go`
  - `full_router_teacher_state_matrix_test.go`
  - `full_router_admin_state_matrix_test.go`
  - `full_router_awd_state_matrix_test.go`

### Boundary choice

- 不创建新的 helper owner，也不把 helper 再拆一轮。
- 只移动顶层测试函数，不修改函数内容。
- 仍然使用 `package app`，让现有 test env 和 helper 无缝共享。

## Task slices

### Slice 1：intake 与计划门禁

- Goal：补齐 reuse decision 和 implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-internal-app-full-router-state-matrix-test-split`

### Slice 2：按领域拆分状态矩阵测试

- Goal：把 contest / teacher / admin / awd 测试迁移到独立文件，原文件只保留 wiring、report preview 和 helper。
- Touched files：
  - `code/backend/internal/app/full_router_state_matrix_integration_test.go`
  - `code/backend/internal/app/full_router_contest_state_matrix_test.go`
  - `code/backend/internal/app/full_router_teacher_state_matrix_test.go`
  - `code/backend/internal/app/full_router_admin_state_matrix_test.go`
  - `code/backend/internal/app/full_router_awd_state_matrix_test.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestTeacherRoutesAreServedByTeachingQuery|TestStudentPracticeReadRoutesAreServedByPracticeModule|TestFullRouter_ReportPreviewAndDownloadStateMatrix|TestFullRouter_ContestParticipationStateMatrix|TestFullRouter_ContestAndReviewArchiveExportStateMatrix|TestFullRouter_TeacherAWDReviewExportStateMatrix|TestFullRouter_TeacherAccessAndRecommendationStateMatrix|TestFullRouter_AdminChallengeManagementStateMatrix|TestFullRouter_InstanceHintAndProxyStateMatrix|TestFullRouter_AWDTrafficAdminStateMatrix|TestFullRouter_ChallengeWriteupsUseCommunitySemantics|TestFullRouter_ContestChallengeAndScoreboardStateMatrix|TestFullRouter_VisibleAWDContestChallengesIncludeAWDServiceID|TestFullRouter_AWDContestLegacyChallengeInstanceRouteRejected|TestFullRouter_AWDChallengeAuthoringStateMatrix|TestFullRouter_AdminContestListSupportsModeStatusesSortAndSummary|TestFullRouter_AdminOpsAndNotificationStateMatrix|TestFullRouter_AdminImagesCapsOversizedPageSize' -count=1`

### Slice 3：一致性与 review 收口

- Goal：确认 harness 文档和拆分后的测试文件没有引入新的结构漂移。
- Validation：
  - `bash scripts/check-consistency.sh`

## Expected change surface

- `internal/app` 集成测试文件组织
- harness reuse decision 与 implementation plan

## Data / API / compatibility impact

- 无生产数据、API、权限或业务语义变更。
- 风险主要在：
  - 移动函数时漏掉 import
  - 原文件删除区间不准确导致重复定义或缺失测试
  - 新文件分组不稳定，后续又回流到一个文件

## Review fit check

- Owner 清晰：原文件收口为 helper owner，新文件各自只承载单一测试领域。
- Reuse 清晰：现有 test env、fixture builder 和 report helper 全部复用，不新造抽象。
- 结构收敛：既然本轮已经触碰 `full_router_state_matrix_integration_test.go` 这个 oversized surface，就直接把顶层测试 owner 拆开，不再继续往同一文件累加。

## Rollback / recovery

- 纯测试代码与文档组织调整，可直接回退到拆分前的单文件版本。
