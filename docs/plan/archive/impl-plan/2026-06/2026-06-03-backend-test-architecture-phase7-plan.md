# 2026-06-03 backend test architecture phase 7 plan

## Objective

- 把 `full_router_awd_state_matrix` 的 HTTP 场景断言 owner 迁到 `code/backend/tests/system/http`。
- 让 `internal/app/full_router_awd_state_matrix_test.go` 只保留 AWD / contest seed、少量数据库断言和最小 glue code。
- 继续沿用前几轮已验证的“先迁断言 owner，后拆 fixture owner”节奏。

## Non-goals

- 本轮不抽取 `newFullRouterTestEnv` 到共享 testutil。
- 本轮不迁移 `full_router_teacher_state_matrix`、`full_router_contest_state_matrix`、`full_router_admin_state_matrix`。
- 本轮不修既有 `TestFullRouter_TeacherAWDReviewExportStateMatrix` 的 PDF 断言失败。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_awd_state_matrix_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouterteacherauthoring/authoring_flow.go`
- `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
- `.harness/reuse-decisions/backend-test-architecture-phase7.md`

## Current problem

- `full_router_awd_state_matrix_test.go` 仍把 AWD 相关五组 HTTP 场景断言放在 `internal/app`。
- 文件混合了 hint/proxy 流程、AWD traffic 管理态、可见题目元数据、legacy route 拒绝和 AWD authoring 状态矩阵，owner 过宽。
- 但如果直接抽 AWD fixture / traffic proxy helper，会把 scope 扩到 contest/runtime 测试基建层。

## Working design

### Target structure

- 新增 `code/backend/tests/system/http/fullrouterawdstate/`：
  - `VerifyInstanceHintAndProxyStateMatrix`
  - `VerifyAWDTrafficAdminStateMatrix`
  - `VerifyVisibleAWDContestChallengesIncludeAWDServiceID`
  - `VerifyAWDContestLegacyChallengeInstanceRouteRejected`
  - `VerifyAWDChallengeAuthoringStateMatrix`
- 新 package 只持有 HTTP 请求序列、断言和小型 driver 定义。
- `internal/app/full_router_awd_state_matrix_test.go` 继续负责：
  - 构造 `newFullRouterTestEnv`
  - AWD contest/team/service seed
  - 少量数据库计数校验与 callback
  - 将 request helper、headers、IDs 和 callback 桥接给新 package

### Boundary choice

- 断言 owner 先迁，AWD seed / proxy target 注入 callback 暂不迁。
- 通过 callback/driver 复用现有 `internal/app` helper，而不是复制 full router fixture。
- 等 AWD 场景迁完，再判断是否值得把 AWD contest/service seed 抽到 testkit。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 7 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase7`

### Slice 2：迁移 full router AWD 场景断言 owner

- Goal：把 `full_router_awd_state_matrix_test.go` 的断言逻辑迁到 `tests/system/http/fullrouterawdstate`。
- Touched files：
  - `code/backend/internal/app/full_router_awd_state_matrix_test.go`
  - `code/backend/tests/system/http/fullrouterawdstate/*.go`
  - `code/backend/tests/README.md`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_InstanceHintAndProxyStateMatrix|TestFullRouter_AWDTrafficAdminStateMatrix|TestFullRouter_VisibleAWDContestChallengesIncludeAWDServiceID|TestFullRouter_AWDContestLegacyChallengeInstanceRouteRejected|TestFullRouter_AWDChallengeAuthoringStateMatrix' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 practice/router 相关现有行为和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- full router AWD state matrix 场景测试 owner
- `tests/system/http` 下第五个专题目录

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - proxy / ticket / cookie 流程抽到新 package 后出现行为漂移
  - legacy route 的 envelope 断言与数据库计数校验拆分后不一致
  - AWD seed 与新断言 package 的边界抽象不清晰

## Review fit check

- Owner 清晰：AWD HTTP 场景断言迁到 `tests/system/http/fullrouterawdstate`。
- Reuse 清晰：继续复用现有 full router fixture，不在这一刀复制 env。
- 结构收敛：这刀只解决 AWD state matrix 场景 owner，不假装一次解决 full router 全部测试基建。

## Rollback / recovery

- 纯测试断言 owner 调整，可直接回退到当前 `internal/app/full_router_awd_state_matrix_test.go` 版本。
