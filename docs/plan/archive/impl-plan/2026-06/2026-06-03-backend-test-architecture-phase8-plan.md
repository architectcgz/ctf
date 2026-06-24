# 2026-06-03 backend test architecture phase 8 plan

## Objective

- 把 `full_router_contest_state_matrix` 的 HTTP 场景断言 owner 迁到 `code/backend/tests/system/http`。
- 让 `internal/app/full_router_contest_state_matrix_test.go` 只保留 contest/challenge/user seed、少量数据库操作和最小 glue code。
- 继续沿用前几轮已验证的“先迁断言 owner，后拆 fixture owner”节奏。

## Non-goals

- 本轮不抽取 `newFullRouterTestEnv` 到共享 testutil。
- 本轮不迁移 `full_router_teacher_state_matrix`、`full_router_state_matrix_integration_test.go`、`full_router_admin_state_matrix_test.go`。
- 本轮不修既有 `TestFullRouter_TeacherAWDReviewExportStateMatrix` 的 PDF 断言失败。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_contest_state_matrix_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouterawdstate/awd_state.go`
- `code/backend/tests/system/http/fullrouterteacherauthoring/authoring_flow.go`
- `.harness/reuse-decisions/backend-test-architecture-phase8.md`

## Current problem

- `full_router_contest_state_matrix_test.go` 仍把 contest participation、export/review archive、challenge/scoreboard、admin contest list 四组 HTTP 场景断言放在 `internal/app`。
- 文件同时持有较长的 HTTP 状态流、报表下载断言和一部分 contest/challenge seed，owner 过宽。
- 但如果直接抽 contest fixture / scoreboard seed helper，会立刻把 scope 扩到更大一层的测试基建重组。

## Working design

### Target structure

- 新增 `code/backend/tests/system/http/fullrouterconteststate/`：
  - `VerifyContestParticipationStateMatrix`
  - `VerifyContestAndReviewArchiveExportStateMatrix`
  - `VerifyContestChallengeAndScoreboardStateMatrix`
  - `VerifyAdminContestListSupportsModeStatusesSortAndSummary`
- 新 package 只持有 HTTP 请求序列、断言和小型 driver 定义。
- `internal/app/full_router_contest_state_matrix_test.go` 继续负责：
  - 构造 `newFullRouterTestEnv`
  - 创建 contest/challenge/user/team/score seed
  - 处理少量数据库删除或查询
  - 将 request helper、headers、IDs 和 callback 桥接给新 package

### Boundary choice

- 断言 owner 先迁，contest/challenge/user seed 暂不迁。
- 通过 callback/driver 复用现有 `internal/app` helper，而不是复制 full router fixture。
- 等 contest 场景迁完，再判断是否值得把 scoreboard/report seed 抽到 testkit。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 8 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase8`

### Slice 2：迁移 full router contest 场景断言 owner

- Goal：把 `full_router_contest_state_matrix_test.go` 的断言逻辑迁到 `tests/system/http/fullrouterconteststate`。
- Touched files：
  - `code/backend/internal/app/full_router_contest_state_matrix_test.go`
  - `code/backend/tests/system/http/fullrouterconteststate/*.go`
  - `code/backend/tests/README.md`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_ContestParticipationStateMatrix|TestFullRouter_ContestAndReviewArchiveExportStateMatrix|TestFullRouter_ContestChallengeAndScoreboardStateMatrix|TestFullRouter_AdminContestListSupportsModeStatusesSortAndSummary' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 practice/router 相关现有行为和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- full router contest state matrix 场景测试 owner
- `tests/system/http` 下第六个专题目录

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - export/review archive 的报表等待与下载断言拆分后出现语义漂移
  - scoreboard 冻结/解冻流转的状态断言迁出后遗漏 seed 前提
  - contest registration/team/submission 的 callback 边界抽象不清晰

## Review fit check

- Owner 清晰：contest HTTP 场景断言迁到 `tests/system/http/fullrouterconteststate`。
- Reuse 清晰：继续复用现有 full router fixture，不在这一刀复制 env。
- 结构收敛：这刀只解决 contest state matrix 场景 owner，不假装一次解决 full router 全部测试基建。

## Rollback / recovery

- 纯测试断言 owner 调整，可直接回退到当前 `internal/app/full_router_contest_state_matrix_test.go` 版本。
