# 2026-06-03 backend test architecture phase 6 plan

## Objective

- 把 `full_router_teacher_authoring` 的 HTTP 场景断言 owner 迁到 `code/backend/tests/system/http`。
- 让 `internal/app/full_router_teacher_authoring_integration_test.go` 只保留 DB / 文件系统 seed、数据库校验和最小 glue code。
- 继续沿用前几轮已验证的“先迁断言 owner，后拆 fixture owner”节奏。

## Non-goals

- 本轮不抽取 `newFullRouterTestEnv` 到共享 testutil。
- 本轮不迁移 `full_router_teacher_state_matrix`、`full_router_contest_*`、`full_router_awd_*`。
- 本轮不修既有 `TestFullRouter_TeacherAWDReviewExportStateMatrix` 的 PDF 断言失败。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_teacher_authoring_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouteraccess/access_matrix.go`
- `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
- `.harness/reuse-decisions/backend-test-architecture-phase6.md`

## Current problem

- `full_router_teacher_authoring_integration_test.go` 仍把 teacher authoring 的三组 HTTP 场景断言放在 `internal/app`。
- 文件同时持有长流程 HTTP 断言、包导出下载断言和少量 DB 结果校验，owner 混在一起，后续还会继续长。
- 但如果直接抽 full router fixture 或通用 challenge seed，会立刻把 scope 扩到更大一层的测试基建重做。

## Working design

### Target structure

- 新增 `code/backend/tests/system/http/fullrouterteacherauthoring/`：
  - `VerifyTeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges`
  - `VerifyCreateChallengeStoresCreatorResponse`
  - `VerifyChallengeSelfCheckRunsPrecheckAndRuntime`
- 新 package 只持有 HTTP 请求序列、断言和小型 driver 定义。
- `internal/app/full_router_teacher_authoring_integration_test.go` 继续负责：
  - 构造 `newFullRouterTestEnv`
  - 包导出 revision 的文件系统 / DB seed
  - `created_by` 的数据库断言
  - 将 request helper、headers 和 callback 桥接给新 package

### Boundary choice

- 断言 owner 先迁，DB / 文件系统 seed 暂不迁。
- 通过 callback/driver 复用 `internal/app` 现有 request helper，而不是复制 full router fixture。
- 等 teacher authoring 场景迁完，再判断是否值得把 authoring challenge seed 抽成 testkit。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 6 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase6`

### Slice 2：迁移 full router teacher authoring 场景断言 owner

- Goal：把 `full_router_teacher_authoring_integration_test.go` 的断言逻辑迁到 `tests/system/http/fullrouterteacherauthoring`。
- Touched files：
  - `code/backend/internal/app/full_router_teacher_authoring_integration_test.go`
  - `code/backend/tests/system/http/fullrouterteacherauthoring/*.go`
  - `code/backend/tests/README.md`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_TeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges|TestFullRouter_CreateChallengeStoresCreator|TestFullRouter_ChallengeSelfCheckRunsPrecheckAndRuntime' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 practice/router 相关现有行为和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- full router teacher authoring 场景测试 owner
- `tests/system/http` 下第四个专题目录

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - 包导出 revision 的 DB / 文件系统 seed 和新断言 package 的边界不清楚
  - HTTP 场景迁出后仍残留过多 glue code，导致价值不足
  - `created_by` 的响应断言和数据库断言拆分后出现语义漂移

## Review fit check

- Owner 清晰：teacher authoring HTTP 场景断言迁到 `tests/system/http/fullrouterteacherauthoring`。
- Reuse 清晰：继续复用现有 full router fixture，不在这一刀复制 env。
- 结构收敛：这刀只解决 teacher authoring 场景 owner，不假装一次解决 full router 全部测试基建。

## Rollback / recovery

- 纯测试断言 owner 调整，可直接回退到当前 `internal/app/full_router_teacher_authoring_integration_test.go` 版本。
