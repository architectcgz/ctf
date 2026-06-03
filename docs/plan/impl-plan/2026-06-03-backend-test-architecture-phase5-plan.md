# 2026-06-03 backend test architecture phase 5 plan

## Objective

- 把 `full_router_admin` 的 HTTP 场景断言 owner 迁到 `code/backend/tests/system/http`。
- 让 `internal/app/full_router_admin_integration_test.go` 只保留 seed、fixture glue 和兼容入口。
- 继续沿用前几轮已验证的“先迁断言 owner，后拆 fixture owner”节奏。

## Non-goals

- 本轮不抽取 `newFullRouterTestEnv` 到共享 testutil。
- 本轮不迁移 `full_router_teacher_*`、`full_router_contest_*`、`full_router_awd_*`。
- 本轮不修既有 `TestFullRouter_TeacherAWDReviewExportStateMatrix` 的 PDF 断言失败。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_admin_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouteraccess/access_matrix.go`
- `.harness/reuse-decisions/backend-test-architecture-phase5.md`

## Current problem

- `full_router_admin_integration_test.go` 仍把两组 admin 场景断言放在 `internal/app`。
- 场景里既有少量 DB seed，也有完整 HTTP 生命周期断言，文件继续增长会让 owner 不清楚。
- 但如果直接抽 full router fixture，会立刻把 scope 扩到大体量 env/seed 共享，不适合这一刀。

## Working design

### Target structure

- 新增 `code/backend/tests/system/http/fullrouteradmin/`：
  - `VerifyAdminCanToggleAWDControlsAndSeeOrchestrationState`
  - `VerifyAdminChallengePublishRequestLifecycle`
- 新 package 只持有 HTTP 请求序列、断言和少量 DTO/driver 定义。
- `internal/app/full_router_admin_integration_test.go` 继续负责：
  - 构造 `newFullRouterTestEnv`
  - 场景前置 DB seed
  - 将 request helper、headers、ID 信息桥接给新 package

### Boundary choice

- 断言 owner 先迁，场景 seed 暂不迁。
- 通过 callback/driver 复用 `internal/app` 现有 request helper，而不是复制 full router fixture。
- 等 admin 场景迁完，再判断是否值得把 full router 场景 seed 抽到 testkit。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 5 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase5`

### Slice 2：迁移 full router admin 场景断言 owner

- Goal：把 `full_router_admin_integration_test.go` 的断言逻辑迁到 `tests/system/http/fullrouteradmin`。
- Touched files：
  - `code/backend/internal/app/full_router_admin_integration_test.go`
  - `code/backend/tests/system/http/fullrouteradmin/*.go`
  - `code/backend/tests/README.md`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_AdminCanToggleAWDControlsAndSeeOrchestrationState|TestFullRouter_AdminChallengePublishRequestLifecycle' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 full router 相关现有行为和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- full router admin 场景测试 owner
- `tests/system/http` 下第三个专题目录

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - 场景 seed 与新断言 package 的职责边界不清晰
  - AWD control / publish request 断言抽象过度，反而降低可读性
  - glue code 漏传路径参数或 headers，导致测试行为漂移

## Review fit check

- Owner 清晰：admin HTTP 场景断言迁到 `tests/system/http/fullrouteradmin`。
- Reuse 清晰：继续复用现有 full router fixture，不在这一刀复制 env。
- 结构收敛：这刀只解决 admin 场景 owner，不假装一次解决 full router 全部测试基建。

## Rollback / recovery

- 纯测试断言 owner 调整，可直接回退到当前 `internal/app/full_router_admin_integration_test.go` 版本。
