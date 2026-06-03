# 2026-06-03 backend test architecture phase 9 plan

## Objective

- 把 `full_router_teacher_state_matrix` 中非 AWD review 的 HTTP 场景断言 owner 迁到 `code/backend/tests/system/http`。
- 让 `internal/app/full_router_teacher_state_matrix_test.go` 只保留 `TeacherAWDReviewExportStateMatrix`、少量 seed 和最小 glue code。
- 继续沿用前几轮已验证的“先迁断言 owner，后拆 fixture owner”节奏。

## Non-goals

- 本轮不抽取 `newFullRouterTestEnv` 到共享 testutil。
- 本轮不迁移 `TestFullRouter_TeacherAWDReviewExportStateMatrix`。
- 本轮不修 `TestFullRouter_TeacherAWDReviewExportStateMatrix` 的既有 PDF 断言失败。
- 本轮不迁移 `full_router_state_matrix_integration_test.go`、`full_router_admin_state_matrix_test.go`。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_teacher_state_matrix_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouterconteststate/contest_state.go`
- `code/backend/tests/system/http/fullrouterteacherauthoring/authoring_flow.go`
- `.harness/reuse-decisions/backend-test-architecture-phase9.md`

## Current problem

- `full_router_teacher_state_matrix_test.go` 仍把 teacher access / recommendation 和 challenge writeup semantics 两组 HTTP 场景断言放在 `internal/app`。
- 同文件还混着 `TeacherAWDReviewExportStateMatrix`，而这个测试有既有失败，导致不能把整个文件一刀迁完。
- 如果把既有失败和架构迁移混做，本轮会失去切片清晰度，也会放大 review 和验证噪声。

## Working design

### Target structure

- 新增 `code/backend/tests/system/http/fullrouterteacherstate/`：
  - `VerifyTeacherAccessAndRecommendationStateMatrix`
  - `VerifyChallengeWriteupsUseCommunitySemantics`
- 新 package 只持有 HTTP 请求序列、断言和小型 driver 定义。
- `internal/app/full_router_teacher_state_matrix_test.go` 继续负责：
  - 保留 `TeacherAWDReviewExportStateMatrix`
  - 构造 `newFullRouterTestEnv`
  - 创建 recommendation/writeup 相关 seed
  - 将 request helper、headers、IDs 和 callback 桥接给新 package

### Boundary choice

- 断言 owner 先迁，teacher/assessment/writeup seed 暂不迁。
- 已知失败的 AWD review export 测试明确排除出本轮。
- 通过 callback/driver 复用现有 `internal/app` helper，而不是复制 full router fixture。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 9 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase9`

### Slice 2：迁移 teacher 非 AWD review 场景断言 owner

- Goal：把 `TeacherAccessAndRecommendationStateMatrix` 和 `ChallengeWriteupsUseCommunitySemantics` 迁到 `tests/system/http/fullrouterteacherstate`。
- Touched files：
  - `code/backend/internal/app/full_router_teacher_state_matrix_test.go`
  - `code/backend/tests/system/http/fullrouterteacherstate/*.go`
  - `code/backend/tests/README.md`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_TeacherAccessAndRecommendationStateMatrix|TestFullRouter_ChallengeWriteupsUseCommunitySemantics' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 practice/router 相关现有行为和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- full router teacher state 中非 AWD review 的场景测试 owner
- `tests/system/http` 下第七个专题目录

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - recommendation / skill-profile / class summary 的权限断言迁出后遗漏角色差异
  - challenge writeup community semantics 的状态流转拆分后出现语义漂移
  - `TeacherAWDReviewExportStateMatrix` 暂留本地导致文件仍是混合 owner，但这是本轮有意保留的边界

## Review fit check

- Owner 清晰：teacher 非 AWD review 的 HTTP 场景断言迁到 `tests/system/http/fullrouterteacherstate`。
- Reuse 清晰：继续复用现有 full router fixture，不在这一刀复制 env。
- 结构收敛：这刀只解决可安全迁出的 teacher state owner，不把既有失败混进来。

## Rollback / recovery

- 纯测试断言 owner 调整，可直接回退到当前 `internal/app/full_router_teacher_state_matrix_test.go` 版本。
